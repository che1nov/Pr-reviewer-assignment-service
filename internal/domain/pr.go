package domain

import "time"

type PullRequest struct {
	ID        string
	Title     string
	AuthorID  string
	TeamName  string
	Reviewers []string
	Status    string
	CreatedAt time.Time
	MergedAt  *time.Time
}

func NewPullRequest(id, title, authorID, teamName string, createdAt time.Time) PullRequest {
	return PullRequest{
		ID:        id,
		Title:     title,
		AuthorID:  authorID,
		TeamName:  teamName,
		Status:    "OPEN",
		Reviewers: make([]string, 0, 2),
		CreatedAt: createdAt,
	}
}

func (pr *PullRequest) AssignReviewers(reviewers []string) {
	pr.Reviewers = reviewers
}

func (pr *PullRequest) MarkMerged(mergedAt time.Time) {
	if pr.Status == "MERGED" {
		if pr.MergedAt == nil {
			pr.MergedAt = &mergedAt
		}
		return
	}
	pr.Status = "MERGED"
	pr.MergedAt = &mergedAt
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
	return nil
}

func (pr *PullRequest) ReplaceReviewer(oldReviewerID, newReviewerID string) error {
	if pr.Status == "MERGED" {
		return ErrPullRequestMerged
	}
	if newReviewerID == pr.AuthorID {
		return ErrReviewerIsAuthor
	}
	if newReviewerID == oldReviewerID {
		return ErrReviewerAlreadyAdded
	}
	for _, existing := range pr.Reviewers {
		if existing == newReviewerID {
			return ErrReviewerAlreadyAdded
		}
	}
	for i, existing := range pr.Reviewers {
		if existing == oldReviewerID {
			pr.Reviewers[i] = newReviewerID
			return nil
		}
	}
	return ErrReviewerNotAssigned
}
