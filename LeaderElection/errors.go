package leaderelection

const (
	// ElectionCompletedNotify indicates the election is over and no new candidates can be nominated.
	ElectionCompletedNotify = "Election is over, no new candidates allowed"
	// ElectionSelfDltNotify indicates that the referenced candidate was deleted. This can happen if the election is
	// ended (EndElection) or deleted (DeleteElection)
	ElectionSelfDltNotify = "Candidate has been deleted"
	// While this candidate was finding who to follow (because it wasn't leader), it became leader. This notification
	// identifies this condition.
	leaderWhileDeterminingWhoToFollow = `While determining who to follow the candidate became leader
	because of a race condition whereby the original leader was deleted after the candidate decided it wasn't leader.`
)