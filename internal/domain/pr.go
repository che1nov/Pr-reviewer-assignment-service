package domain

type PullRequest struct {
	ID                string
	Title             string
	AuthorID          string
	TeamName          string
	Reviewers         []string
	Status            string
	NeedMoreReviewers bool
}

func NewPullRequest(id, title, authorID, teamName string) PullRequest {
	return PullRequest{
		ID:                id,
		Title:             title,
		AuthorID:          authorID,
		TeamName:          teamName,
		Status:            "OPEN",
		Reviewers:         []string{},
		NeedMoreReviewers: true,
	}
}

func (pr *PullRequest) AssignReviewers(reviewers []string) {
	pr.Reviewers = reviewers
	pr.NeedMoreReviewers = len(reviewers) < 2
}

func (pr *PullRequest) MarkMerged() {
	pr.Status = "MERGED"
}

func (pr *PullRequest) AddReviewer(reviewerID string) error {
	if pr.Status == "MERGED" {
		return ErrPullRequestMerged
	}
	if reviewerID == pr.AuthorID {
		return ErrReviewerIsAuthor
	}
	for _, existing := range pr.Reviewers {
		if existing == reviewerID {
			return ErrReviewerAlreadyAdded
		}
	}
	if len(pr.Reviewers) >= 2 {
		return ErrReviewerLimitReached
	}
	pr.Reviewers = append(pr.Reviewers, reviewerID)
	pr.NeedMoreReviewers = len(pr.Reviewers) < 2
	return nil
}

func (pr *PullRequest) ReplaceReviewer(oldReviewerID, newReviewerID string) error {
	if pr.Status == "MERGED" {
		return ErrPullRequestMerged
	}
	if newReviewerID == pr.AuthorID {
		return ErrReviewerIsAuthor
	}
	found := false
	for _, existing := range pr.Reviewers {
		if existing == newReviewerID {
			return ErrReviewerAlreadyAdded
		}
	}
	for i, existing := range pr.Reviewers {
		if existing == oldReviewerID {
			pr.Reviewers[i] = newReviewerID
			found = true
			break
		}
	}
	if !found {
		return ErrReviewerNotAssigned
	}
	pr.NeedMoreReviewers = len(pr.Reviewers) < 2
	return nil
}
