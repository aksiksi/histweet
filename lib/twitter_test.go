package histweet

import (
	"net/http"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
)

// Mock Twitter API services for testing
type mockTwitterAccountService struct{}

func (s *mockTwitterAccountService) VerifyCredentials(params *twitter.AccountVerifyParams) (*twitter.User, *http.Response, error) {
	return nil, nil, nil
}

type mockTwitterTimelineService struct{}

func (s *mockTwitterTimelineService) UserTimeline(params *twitter.UserTimelineParams) ([]twitter.Tweet, *http.Response, error) {
	tweets := make([]twitter.Tweet, 100)

	tweets[0].FavoriteCount = 10
	tweets[1].Text = "potato"

	return tweets, nil, nil
}

type mockTwitterStatusService struct{}

func (s *mockTwitterStatusService) Destroy(id int64, params *twitter.StatusDestroyParams) (*twitter.Tweet, *http.Response, error) {
	return nil, nil, nil
}

type mockTwitterClient struct{}

func (t *mockTwitterClient) accountService() twitterAccountService {
	return &mockTwitterAccountService{}
}

func (t *mockTwitterClient) timelineService() twitterTimelineService {
	return &mockTwitterTimelineService{}
}

func (t *mockTwitterClient) statusService() twitterStatusService {
	return &mockTwitterStatusService{}
}

func TestNewTwitterClient(t *testing.T) {
	// The first call will fail due to bad OAuth creds, but we can ignore the error
	_, _ = NewTwitterClient("", "", "", "", true)

	_, err := NewTwitterClient("", "", "", "", false)
	if err != nil {
		t.Error(err)
	}
}

func TestTweetAPIs(t *testing.T) {
	client := &mockTwitterClient{}

	tweetRule, _ := Parse(`likes >= 3 || text ~ "potato"`)

	// Test cases
	var inputs = []struct {
		name    string
		rule    *Rule
		matches int
	}{
		{"rule_tweet", &Rule{Tweet: tweetRule}, 2},

		// The count rule will keep the 10 latest tweets
		// Since we have 100 tweets total (above), 90 will be deleted
		{"rule_count", &Rule{Count: &RuleCount{N: 10}}, 90},
	}

	for _, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			tweets, _ := FetchTimelineTweets(input.rule, client)
			if len(tweets) != input.matches {
				t.Errorf("Expected %d tweets to match the rule", input.matches)
			}

			DeleteTweets(tweets, client)
		})
	}
}
