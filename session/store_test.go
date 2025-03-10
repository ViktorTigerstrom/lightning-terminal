package session

import (
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/lightningnetwork/lnd/clock"
	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// TestBasicSessionStore tests the basic getters and setters of the session
// store.
func TestBasicSessionStore(t *testing.T) {
	// Set up a new DB.
	clock := clock.NewTestClock(testTime)
	db, err := NewDB(t.TempDir(), "test.db", clock)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	// Create a few sessions. We increment the time by one second between
	// each session to ensure that the created at time is unique and hence
	// that the ListSessions method returns the sessions in a deterministic
	// order.
	s1 := newSession(t, db, clock, "session 1")
	clock.SetTime(testTime.Add(time.Second))
	s2 := newSession(t, db, clock, "session 2")
	clock.SetTime(testTime.Add(2 * time.Second))
	s3 := newSession(t, db, clock, "session 3", withType(TypeAutopilot))
	clock.SetTime(testTime.Add(3 * time.Second))
	s4 := newSession(t, db, clock, "session 4")

	// Persist session 1. This should now succeed.
	require.NoError(t, db.CreateSession(s1))

	// Trying to persist session 1 again should fail due to a session with
	// the given pub key already existing.
	require.ErrorContains(t, db.CreateSession(s1), "already exists")

	// Change the local pub key of session 4 such that it has the same
	// ID as session 1.
	s4.ID = s1.ID
	s4.GroupID = s1.GroupID

	// Now try to insert session 4. This should fail due to an entry for
	// the ID already existing.
	require.ErrorContains(t, db.CreateSession(s4), "a session with the "+
		"given ID already exists")

	// Persist a few more sessions.
	require.NoError(t, db.CreateSession(s2))
	require.NoError(t, db.CreateSession(s3))

	// Test the ListSessionsByType method.
	sessions, err := db.ListSessionsByType(TypeMacaroonAdmin)
	require.NoError(t, err)
	require.Equal(t, 2, len(sessions))
	assertEqualSessions(t, s1, sessions[0])
	assertEqualSessions(t, s2, sessions[1])

	sessions, err = db.ListSessionsByType(TypeAutopilot)
	require.NoError(t, err)
	require.Equal(t, 1, len(sessions))
	assertEqualSessions(t, s3, sessions[0])

	sessions, err = db.ListSessionsByType(TypeMacaroonReadonly)
	require.NoError(t, err)
	require.Empty(t, sessions)

	// Ensure that we can retrieve each session by both its local pub key
	// and by its ID.
	for _, s := range []*Session{s1, s2, s3} {
		session, err := db.GetSession(s.LocalPublicKey)
		require.NoError(t, err)
		assertEqualSessions(t, s, session)

		session, err = db.GetSessionByID(s.ID)
		require.NoError(t, err)
		assertEqualSessions(t, s, session)
	}

	// Fetch session 1 and assert that it currently has no remote pub key.
	session1, err := db.GetSession(s1.LocalPublicKey)
	require.NoError(t, err)
	require.Nil(t, session1.RemotePublicKey)

	// Use the update method to add a remote key.
	remotePriv, err := btcec.NewPrivateKey()
	require.NoError(t, err)
	remotePub := remotePriv.PubKey()

	err = db.UpdateSessionRemotePubKey(session1.LocalPublicKey, remotePub)
	require.NoError(t, err)

	// Assert that the session now does have the remote pub key.
	session1, err = db.GetSession(s1.LocalPublicKey)
	require.NoError(t, err)
	require.True(t, remotePub.IsEqual(session1.RemotePublicKey))

	// Check that the session's state is currently StateCreated.
	require.Equal(t, session1.State, StateCreated)

	// Now revoke the session and assert that the state is revoked.
	require.NoError(t, db.ShiftState(s1.ID, StateRevoked))
	s1, err = db.GetSession(s1.LocalPublicKey)
	require.NoError(t, err)
	require.Equal(t, s1.State, StateRevoked)

	// Test that ListAllSessions works.
	sessions, err = db.ListAllSessions()
	require.NoError(t, err)
	require.Equal(t, 3, len(sessions))
	assertEqualSessions(t, s1, sessions[0])
	assertEqualSessions(t, s2, sessions[1])
	assertEqualSessions(t, s3, sessions[2])

	// Test that ListSessionsByState works.
	sessions, err = db.ListSessionsByState(StateRevoked)
	require.NoError(t, err)
	require.Equal(t, 1, len(sessions))
	assertEqualSessions(t, s1, sessions[0])

	sessions, err = db.ListSessionsByState(StateCreated)
	require.NoError(t, err)
	require.Equal(t, 2, len(sessions))
	assertEqualSessions(t, s2, sessions[0])
	assertEqualSessions(t, s3, sessions[1])

	sessions, err = db.ListSessionsByState(StateCreated, StateRevoked)
	require.NoError(t, err)
	require.Equal(t, 3, len(sessions))
	assertEqualSessions(t, s1, sessions[0])
	assertEqualSessions(t, s2, sessions[1])
	assertEqualSessions(t, s3, sessions[2])

	sessions, err = db.ListSessionsByState()
	require.NoError(t, err)
	require.Empty(t, sessions)

	sessions, err = db.ListSessionsByState(StateReserved)
	require.NoError(t, err)
	require.Empty(t, sessions)

	// Demonstrate deletion of a reserved session.
	//
	// Calling DeleteReservedSessions should have no effect yet since none
	// of the sessions are reserved.
	require.NoError(t, db.DeleteReservedSessions())

	sessions, err = db.ListSessionsByState(StateReserved)
	require.NoError(t, err)
	require.Empty(t, sessions)

	// Add a session and put it in the StateReserved state. We'll also
	// link it to session 1.
	s5 := newSession(
		t, db, clock, "session 5", withState(StateReserved),
		withLinkedGroupID(&session1.GroupID),
	)
	require.NoError(t, db.CreateSession(s5))

	sessions, err = db.ListSessionsByState(StateReserved)
	require.NoError(t, err)
	require.Equal(t, 1, len(sessions))
	assertEqualSessions(t, s5, sessions[0])

	// Show that the group ID/session ID index has also been populated with
	// this session.
	groupID, err := db.GetGroupID(s5.ID)
	require.NoError(t, err)
	require.Equal(t, s1.ID, groupID)

	sessIDs, err := db.GetSessionIDs(s5.GroupID)
	require.NoError(t, err)
	require.ElementsMatch(t, []ID{s5.ID, s1.ID}, sessIDs)

	// Now delete the reserved session and show that it is no longer in the
	// database and no longer in the group ID/session ID index.
	require.NoError(t, db.DeleteReservedSessions())

	sessions, err = db.ListSessionsByState(StateReserved)
	require.NoError(t, err)
	require.Empty(t, sessions)

	_, err = db.GetGroupID(s5.ID)
	require.ErrorContains(t, err, "no index entry")

	// Only session 1 should remain in this group.
	sessIDs, err = db.GetSessionIDs(s5.GroupID)
	require.NoError(t, err)
	require.ElementsMatch(t, []ID{s1.ID}, sessIDs)
}

// TestLinkingSessions tests that session linking works as expected.
func TestLinkingSessions(t *testing.T) {
	// Set up a new DB.
	clock := clock.NewTestClock(testTime)
	db, err := NewDB(t.TempDir(), "test.db", clock)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	// Create a new session with no previous link.
	s1 := newSession(t, db, clock, "session 1")

	// Create another session and link it to the first.
	s2 := newSession(t, db, clock, "session 2", withLinkedGroupID(&s1.GroupID))

	// Try to persist the second session and assert that it fails due to the
	// linked session not existing in the DB yet.
	require.ErrorContains(t, db.CreateSession(s2), "unknown linked session")

	// Now persist the first session and retry persisting the second one
	// and assert that this now works.
	require.NoError(t, db.CreateSession(s1))

	// Persisting the second session immediately should fail due to the
	// first session still being active.
	require.ErrorContains(t, db.CreateSession(s2), "is still active")

	// Revoke the first session.
	require.NoError(t, db.ShiftState(s1.ID, StateRevoked))

	// Persisting the second linked session should now work.
	require.NoError(t, db.CreateSession(s2))
}

// TestIDToGroupIDIndex tests that the session-ID-to-group-ID and
// group-ID-to-session-ID indexes work as expected by asserting the behaviour
// of the GetGroupID and GetSessionIDs methods.
func TestLinkedSessions(t *testing.T) {
	// Set up a new DB.
	clock := clock.NewTestClock(testTime)
	db, err := NewDB(t.TempDir(), "test.db", clock)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	// Create a few sessions. The first one is a new session and the two
	// after are all linked to the prior one. All these sessions belong to
	// the same group. The group ID is equivalent to the session ID of the
	// first session.
	s1 := newSession(t, db, clock, "session 1")
	s2 := newSession(
		t, db, clock, "session 2", withLinkedGroupID(&s1.GroupID),
	)
	s3 := newSession(
		t, db, clock, "session 3", withLinkedGroupID(&s2.GroupID),
	)

	// Persist the sessions.
	require.NoError(t, db.CreateSession(s1))

	require.NoError(t, db.ShiftState(s1.ID, StateRevoked))
	require.NoError(t, db.CreateSession(s2))

	require.NoError(t, db.ShiftState(s2.ID, StateRevoked))
	require.NoError(t, db.CreateSession(s3))

	// Assert that the session ID to group ID index works as expected.
	for _, s := range []*Session{s1, s2, s3} {
		groupID, err := db.GetGroupID(s.ID)
		require.NoError(t, err)
		require.Equal(t, s1.ID, groupID)
		require.Equal(t, s.GroupID, groupID)
	}

	// Assert that the group ID to session ID index works as expected.
	sIDs, err := db.GetSessionIDs(s1.GroupID)
	require.NoError(t, err)
	require.EqualValues(t, []ID{s1.ID, s2.ID, s3.ID}, sIDs)

	// To ensure that different groups don't interfere with each other,
	// let's add another set of linked sessions not linked to the first.
	s4 := newSession(t, db, clock, "session 4")
	s5 := newSession(t, db, clock, "session 5", withLinkedGroupID(&s4.GroupID))

	require.NotEqual(t, s4.GroupID, s1.GroupID)

	// Persist the sessions.
	require.NoError(t, db.CreateSession(s4))
	require.NoError(t, db.ShiftState(s4.ID, StateRevoked))

	require.NoError(t, db.CreateSession(s5))

	// Assert that the session ID to group ID index works as expected.
	for _, s := range []*Session{s4, s5} {
		groupID, err := db.GetGroupID(s.ID)
		require.NoError(t, err)
		require.Equal(t, s4.ID, groupID)
		require.Equal(t, s.GroupID, groupID)
	}

	// Assert that the group ID to session ID index works as expected.
	sIDs, err = db.GetSessionIDs(s5.GroupID)
	require.NoError(t, err)
	require.EqualValues(t, []ID{s4.ID, s5.ID}, sIDs)
}

// TestCheckSessionGroupPredicate asserts that the CheckSessionGroupPredicate
// method correctly checks if each session in a group passes a predicate.
func TestCheckSessionGroupPredicate(t *testing.T) {
	// Set up a new DB.
	clock := clock.NewTestClock(testTime)
	db, err := NewDB(t.TempDir(), "test.db", clock)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	// We will use the Label of the Session to test that the predicate
	// function is checked correctly.

	// Add a new session to the DB.
	s1 := newSession(t, db, clock, "label 1")
	require.NoError(t, db.CreateSession(s1))

	// Check that the group passes against an appropriate predicate.
	ok, err := db.CheckSessionGroupPredicate(
		s1.GroupID, func(s *Session) bool {
			return strings.Contains(s.Label, "label 1")
		},
	)
	require.NoError(t, err)
	require.True(t, ok)

	// Check that the group fails against an appropriate predicate.
	ok, err = db.CheckSessionGroupPredicate(
		s1.GroupID, func(s *Session) bool {
			return strings.Contains(s.Label, "label 2")
		},
	)
	require.NoError(t, err)
	require.False(t, ok)

	// Revoke the first session.
	require.NoError(t, db.ShiftState(s1.ID, StateRevoked))

	// Add a new session to the same group as the first one.
	s2 := newSession(t, db, clock, "label 2", withLinkedGroupID(&s1.GroupID))
	require.NoError(t, db.CreateSession(s2))

	// Check that the group passes against an appropriate predicate.
	ok, err = db.CheckSessionGroupPredicate(
		s1.GroupID, func(s *Session) bool {
			return strings.Contains(s.Label, "label")
		},
	)
	require.NoError(t, err)
	require.True(t, ok)

	// Check that the group fails against an appropriate predicate.
	ok, err = db.CheckSessionGroupPredicate(
		s1.GroupID, func(s *Session) bool {
			return strings.Contains(s.Label, "label 1")
		},
	)
	require.NoError(t, err)
	require.False(t, ok)

	// Add a new session that is not linked to the first one.
	s3 := newSession(t, db, clock, "completely different")
	require.NoError(t, db.CreateSession(s3))

	// Ensure that the first group is unaffected.
	ok, err = db.CheckSessionGroupPredicate(
		s1.GroupID, func(s *Session) bool {
			return strings.Contains(s.Label, "label")
		},
	)
	require.NoError(t, err)
	require.True(t, ok)

	// And that the new session is evaluated separately.
	ok, err = db.CheckSessionGroupPredicate(
		s3.GroupID, func(s *Session) bool {
			return strings.Contains(s.Label, "label")
		},
	)
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = db.CheckSessionGroupPredicate(
		s3.GroupID, func(s *Session) bool {
			return strings.Contains(s.Label, "different")
		},
	)
	require.NoError(t, err)
	require.True(t, ok)
}

// TestStateShift tests that the ShiftState method works as expected.
func TestStateShift(t *testing.T) {
	// Set up a new DB.
	clock := clock.NewTestClock(testTime)
	db, err := NewDB(t.TempDir(), "test.db", clock)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	// Add a new session to the DB.
	s1 := newSession(t, db, clock, "label 1")
	require.NoError(t, db.CreateSession(s1))

	// Check that the session is in the StateCreated state. Also check that
	// the "RevokedAt" time has not yet been set.
	s1, err = db.GetSession(s1.LocalPublicKey)
	require.NoError(t, err)
	require.Equal(t, StateCreated, s1.State)
	require.Equal(t, time.Time{}, s1.RevokedAt)

	// Shift the state of the session to StateRevoked.
	err = db.ShiftState(s1.ID, StateRevoked)
	require.NoError(t, err)

	// This should have worked. Since it is now in a terminal state, the
	// "RevokedAt" time should be set.
	s1, err = db.GetSession(s1.LocalPublicKey)
	require.NoError(t, err)
	require.Equal(t, StateRevoked, s1.State)
	require.True(t, clock.Now().Equal(s1.RevokedAt))

	// Trying to do the same state shift again should succeed since the
	// session is already in the expected "dest" state. The revoked-at time
	// should not have changed though.
	prevTime := clock.Now()
	clock.SetTime(prevTime.Add(time.Second))
	err = db.ShiftState(s1.ID, StateRevoked)
	require.NoError(t, err)
	require.True(t, prevTime.Equal(s1.RevokedAt))

	// Trying to shift the state from a terminal state back to StateCreated
	// should also fail since this is not a legal state transition.
	err = db.ShiftState(s1.ID, StateCreated)
	require.ErrorContains(t, err, "illegal session state transition")
}

// testSessionModifier is a functional option that can be used to modify the
// default test session created by newSession.
type testSessionModifier func(*Session)

func withLinkedGroupID(groupID *ID) testSessionModifier {
	return func(s *Session) {
		s.GroupID = *groupID
	}
}

func withType(t Type) testSessionModifier {
	return func(s *Session) {
		s.Type = t
	}
}

func withState(state State) testSessionModifier {
	return func(s *Session) {
		s.State = state
	}
}

func newSession(t *testing.T, db Store, clock clock.Clock, label string,
	mods ...testSessionModifier) *Session {

	id, priv, err := db.GetUnusedIDAndKeyPair()
	require.NoError(t, err)

	session, err := buildSession(
		id, priv, label, TypeMacaroonAdmin,
		clock.Now(),
		time.Date(99999, 1, 1, 0, 0, 0, 0, time.UTC),
		"foo.bar.baz:1234", true, nil, nil, nil, true, nil,
		[]PrivacyFlag{ClearPubkeys},
	)
	require.NoError(t, err)

	for _, mod := range mods {
		mod(session)
	}

	return session
}

func assertEqualSessions(t *testing.T, expected, actual *Session) {
	expectedExpiry := expected.Expiry
	actualExpiry := actual.Expiry
	expectedRevoked := expected.RevokedAt
	actualRevoked := actual.RevokedAt
	expectedCreated := expected.CreatedAt
	actualCreated := actual.CreatedAt

	expected.Expiry = time.Time{}
	expected.RevokedAt = time.Time{}
	expected.CreatedAt = time.Time{}
	actual.Expiry = time.Time{}
	actual.RevokedAt = time.Time{}
	actual.CreatedAt = time.Time{}

	require.Equal(t, expected, actual)
	require.Equal(t, expectedExpiry.Unix(), actualExpiry.Unix())
	require.Equal(t, expectedRevoked.Unix(), actualRevoked.Unix())
	require.Equal(t, expectedCreated.Unix(), actualCreated.Unix())

	// Restore the old values to not influence the tests.
	expected.Expiry = expectedExpiry
	expected.RevokedAt = expectedRevoked
	expected.CreatedAt = expectedCreated
	actual.Expiry = actualExpiry
	actual.RevokedAt = actualRevoked
	actual.CreatedAt = actualCreated
}
