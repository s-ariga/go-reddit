package geddit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// UserService handles communication with the user
// related methods of the Reddit API
type UserService interface {
	Get(ctx context.Context, username string) (*User, *Response, error)
	GetMultipleByID(ctx context.Context, ids ...string) (map[string]*UserShort, *Response, error)
	UsernameAvailable(ctx context.Context, username string) (bool, *Response, error)

	Overview(ctx context.Context, opts *ListOptions) (*CommentsLinks, *Response, error)
	OverviewOf(ctx context.Context, username string, opts *ListOptions) (*CommentsLinks, *Response, error)

	// returns the client's links
	GetHotLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error)
	GetNewLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error)
	GetTopLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error)
	GetControversialLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error)

	// returns the links of the user with the username
	GetHotLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error)
	GetNewLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error)
	GetTopLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error)
	GetControversialLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error)

	GetUpvoted(ctx context.Context, opts *ListOptions) (*Links, *Response, error)
	GetDownvoted(ctx context.Context, opts *ListOptions) (*Links, *Response, error)
	GetHidden(ctx context.Context, opts *ListOptions) (*Links, *Response, error)
	GetSaved(ctx context.Context, opts *ListOptions) (*CommentsLinks, *Response, error)
	GetGilded(ctx context.Context, opts *ListOptions) (*CommentsLinks, *Response, error)

	// returns the client's comments
	GetHotComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error)
	GetNewComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error)
	GetTopComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error)
	GetControversialComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error)

	// returns the comments of the user with the username
	GetHotCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error)
	GetNewCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error)
	GetTopCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error)
	GetControversialCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error)

	Friend(ctx context.Context, username string, note string) (interface{}, *Response, error)
	Unblock(ctx context.Context, username string) (*Response, error)
	Unfriend(ctx context.Context, username string) (*Response, error)
}

// UserServiceOp implements the UserService interface
type UserServiceOp struct {
	client *Client
}

var _ UserService = &UserServiceOp{}

// User represents a Reddit user
type User struct {
	// is not the full ID, watch out
	ID      string     `json:"id,omitempty"`
	Name    string     `json:"name,omitempty"`
	Created *Timestamp `json:"created_utc,omitempty"`

	LinkKarma    int `json:"link_karma"`
	CommentKarma int `json:"comment_karma"`

	IsFriend         bool `json:"is_friend"`
	IsEmployee       bool `json:"is_employee"`
	HasVerifiedEmail bool `json:"has_verified_email"`
	NSFW             bool `json:"over_18"`
}

// UserShort represents a Reddit user, but contains fewer pieces of information
// It is returned from the GET /api/user_data_by_account_ids endpoint
type UserShort struct {
	Name    string     `json:"name,omitempty"`
	Created *Timestamp `json:"created_utc,omitempty"`

	LinkKarma    int `json:"link_karma"`
	CommentKarma int `json:"comment_karma"`

	NSFW bool `json:"profile_over_18"`
}

// Get returns information about the user
func (s *UserServiceOp) Get(ctx context.Context, username string) (*User, *Response, error) {
	path := fmt.Sprintf("user/%s/about", username)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(userRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Data, resp, nil
}

// GetMultipleByID returns multiple users from their full IDs
// The response body is a map where the keys are the IDs (if they exist), and the value is the user
func (s *UserServiceOp) GetMultipleByID(ctx context.Context, ids ...string) (map[string]*UserShort, *Response, error) {
	type query struct {
		IDs []string `url:"ids,omitempty,comma"`
	}

	path := "api/user_data_by_account_ids"
	path, err := addOptions(path, query{ids})
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(map[string]*UserShort)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return *root, resp, nil
}

// UsernameAvailable checks whether a username is available for registration
// If a valid username is provided, this endpoint returns a body with just "true" or "false"
func (s *UserServiceOp) UsernameAvailable(ctx context.Context, username string) (bool, *Response, error) {
	type query struct {
		User string `url:"user,omitempty"`
	}

	path := "api/username_available"
	path, err := addOptions(path, query{username})
	if err != nil {
		return false, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return false, nil, err
	}

	root := new(bool)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return false, resp, err
	}

	return *root, resp, nil
}

// Overview returns a list of the client's comments and links
func (s *UserServiceOp) Overview(ctx context.Context, opts *ListOptions) (*CommentsLinks, *Response, error) {
	return s.OverviewOf(ctx, s.client.Username, opts)
}

// OverviewOf returns a list of the user's comments and links
func (s *UserServiceOp) OverviewOf(ctx context.Context, username string, opts *ListOptions) (*CommentsLinks, *Response, error) {
	path := fmt.Sprintf("user/%s/overview", username)
	return s.getCommentsAndLinks(ctx, path, opts)
}

// GetHotLinks returns a list of the client's hottest submissions
func (s *UserServiceOp) GetHotLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error) {
	return s.GetHotLinksOf(ctx, s.client.Username, opts)
}

// GetNewLinks returns a list of the client's newest submissions
func (s *UserServiceOp) GetNewLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error) {
	return s.GetNewLinksOf(ctx, s.client.Username, opts)
}

// GetTopLinks returns a list of the client's top submissions
func (s *UserServiceOp) GetTopLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error) {
	return s.GetTopLinksOf(ctx, s.client.Username, opts)
}

// GetControversialLinks returns a list of the client's most controversial submissions
func (s *UserServiceOp) GetControversialLinks(ctx context.Context, opts *ListOptions) (*Links, *Response, error) {
	return s.GetControversialLinksOf(ctx, s.client.Username, opts)
}

// GetHotLinksOf returns a list of the user's hottest submissions
func (s *UserServiceOp) GetHotLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error) {
	path := fmt.Sprintf("user/%s/submitted?sort=%s", username, SortHot)
	return s.getLinks(ctx, path, opts)
}

// GetNewLinksOf returns a list of the user's newest submissions
func (s *UserServiceOp) GetNewLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error) {
	path := fmt.Sprintf("user/%s/submitted?sort=%s", username, SortNew)
	return s.getLinks(ctx, path, opts)
}

// GetTopLinksOf returns a list of the user's top submissions
func (s *UserServiceOp) GetTopLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error) {
	path := fmt.Sprintf("user/%s/submitted?sort=%s", username, SortTop)
	return s.getLinks(ctx, path, opts)
}

// GetControversialLinksOf returns a list of the user's most controversial submissions
func (s *UserServiceOp) GetControversialLinksOf(ctx context.Context, username string, opts *ListOptions) (*Links, *Response, error) {
	path := fmt.Sprintf("user/%s/submitted?sort=%s", username, SortControversial)
	return s.getLinks(ctx, path, opts)
}

// GetUpvoted returns a list of the client's upvoted submissions
func (s *UserServiceOp) GetUpvoted(ctx context.Context, opts *ListOptions) (*Links, *Response, error) {
	path := fmt.Sprintf("user/%s/upvoted", s.client.Username)
	return s.getLinks(ctx, path, opts)
}

// GetDownvoted returns a list of the client's downvoted submissions
func (s *UserServiceOp) GetDownvoted(ctx context.Context, opts *ListOptions) (*Links, *Response, error) {
	path := fmt.Sprintf("user/%s/downvoted", s.client.Username)
	return s.getLinks(ctx, path, opts)
}

// GetHidden returns a list of the client's hidden submissions
func (s *UserServiceOp) GetHidden(ctx context.Context, opts *ListOptions) (*Links, *Response, error) {
	path := fmt.Sprintf("user/%s/hidden", s.client.Username)
	return s.getLinks(ctx, path, opts)
}

// GetSaved returns a list of the client's saved comments and links
func (s *UserServiceOp) GetSaved(ctx context.Context, opts *ListOptions) (*CommentsLinks, *Response, error) {
	path := fmt.Sprintf("user/%s/saved", s.client.Username)
	return s.getCommentsAndLinks(ctx, path, opts)
}

// GetGilded returns a list of the client's gilded comments and links
func (s *UserServiceOp) GetGilded(ctx context.Context, opts *ListOptions) (*CommentsLinks, *Response, error) {
	path := fmt.Sprintf("user/%s/gilded", s.client.Username)
	return s.getCommentsAndLinks(ctx, path, opts)
}

// GetHotComments returns a list of the client's hottest comments
func (s *UserServiceOp) GetHotComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error) {
	return s.GetHotCommentsOf(ctx, s.client.Username, opts)
}

// GetNewComments returns a list of the client's newest comments
func (s *UserServiceOp) GetNewComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error) {
	return s.GetNewCommentsOf(ctx, s.client.Username, opts)
}

// GetTopComments returns a list of the client's top comments
func (s *UserServiceOp) GetTopComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error) {
	return s.GetTopCommentsOf(ctx, s.client.Username, opts)
}

// GetControversialComments returns a list of the client's most controversial comments
func (s *UserServiceOp) GetControversialComments(ctx context.Context, opts *ListOptions) (*Comments, *Response, error) {
	return s.GetControversialCommentsOf(ctx, s.client.Username, opts)
}

// GetHotCommentsOf returns a list of the user's hottest comments
func (s *UserServiceOp) GetHotCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error) {
	path := fmt.Sprintf("user/%s/comments?sort=%s", username, SortHot)
	return s.getComments(ctx, path, opts)
}

// GetNewCommentsOf returns a list of the user's newest comments
func (s *UserServiceOp) GetNewCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error) {
	path := fmt.Sprintf("user/%s/comments?sort=%s", username, SortNew)
	return s.getComments(ctx, path, opts)
}

// GetTopCommentsOf returns a list of the user's top comments
func (s *UserServiceOp) GetTopCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error) {
	path := fmt.Sprintf("user/%s/comments?sort=%s", username, SortTop)
	return s.getComments(ctx, path, opts)
}

// GetControversialCommentsOf returns a list of the user's most controversial comments
func (s *UserServiceOp) GetControversialCommentsOf(ctx context.Context, username string, opts *ListOptions) (*Comments, *Response, error) {
	path := fmt.Sprintf("user/%s/comments?sort=%s", username, SortControversial)
	return s.getComments(ctx, path, opts)
}

// Friend creates or updates a "friend" relationship
// Request body contains JSON data with:
//   name: existing Reddit username
//   note: a string no longer than 300 characters
func (s *UserServiceOp) Friend(ctx context.Context, username string, note string) (interface{}, *Response, error) {
	type request struct {
		Name string `url:"name"`
		Note string `url:"note"`
	}

	path := fmt.Sprintf("api/v1/me/friends/%s", username)
	body := request{Name: username, Note: note}

	_, err := s.client.NewRequest(http.MethodPut, path, body)
	if err != nil {
		return false, nil, err
	}

	// todo: requires gold
	return nil, nil, nil
}

// Unblock unblocks a user
func (s *UserServiceOp) Unblock(ctx context.Context, username string) (*Response, error) {
	path := "api/unfriend"

	form := url.Values{}
	form.Set("name", username)
	form.Set("type", "enemy")
	form.Set("container", "todo: this should be the current user's full id")

	req, err := s.client.NewPostForm(path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Unfriend unfriends a user
func (s *UserServiceOp) Unfriend(ctx context.Context, username string) (*Response, error) {
	path := fmt.Sprintf("api/v1/me/friends/%s", username)
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *UserServiceOp) getLinks(ctx context.Context, path string, opts *ListOptions) (*Links, *Response, error) {
	listing, resp, err := s.getListing(ctx, path, opts)
	if err != nil {
		return nil, resp, err
	}
	return listing.getLinks(), resp, nil
}

func (s *UserServiceOp) getComments(ctx context.Context, path string, opts *ListOptions) (*Comments, *Response, error) {
	listing, resp, err := s.getListing(ctx, path, opts)
	if err != nil {
		return nil, resp, err
	}
	return listing.getComments(), resp, nil
}

func (s *UserServiceOp) getCommentsAndLinks(ctx context.Context, path string, opts *ListOptions) (*CommentsLinks, *Response, error) {
	listing, resp, err := s.getListing(ctx, path, opts)
	if err != nil {
		return nil, resp, err
	}

	v := new(CommentsLinks)
	v.Comments = listing.getComments().Comments
	v.Links = listing.getLinks().Links
	v.After = listing.getAfter()
	v.Before = listing.getBefore()

	return v, resp, nil
}

func (s *UserServiceOp) getListing(ctx context.Context, path string, opts *ListOptions) (*rootListing, *Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(rootListing)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, err
}
