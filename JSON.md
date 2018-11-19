# Reddit JSON Example

I couldn't find a comprehensive reference for the JSON returned by Reddit, so
static parsing caused problems. For instance, if a post does not have images,
the entire "preview" array will be missing.  In some cases (like edited
messages), a new array will appear. This caused too much confusion so the
program was changed to deal with JSON input without a "schema". The original
reddit response (for an `/r/aww/random.json`) is seen below for reference.

```go
// redditPaste holds the return of a reddit article request.
type redditResponse []struct {
	Kind string `json:"kind"`
	Data struct {
		Modhash  string `json:"modhash"`
		Dist     int    `json:"dist"`
		Children []struct {
			Kind string `json:"kind"`
			Data struct {
				ApprovedAtUtc              interface{}   `json:"approved_at_utc"`
				Subreddit                  string        `json:"subreddit"`
				Selftext                   string        `json:"selftext"`
				UserReports                []interface{} `json:"user_reports"`
				Saved                      bool          `json:"saved"`
				ModReasonTitle             interface{}   `json:"mod_reason_title"`
				Gilded                     int           `json:"gilded"`
				Clicked                    bool          `json:"clicked"`
				Title                      string        `json:"title"`
				LinkFlairRichtext          []interface{} `json:"link_flair_richtext"`
				SubredditNamePrefixed      string        `json:"subreddit_name_prefixed"`
				Hidden                     bool          `json:"hidden"`
				Pwls                       int           `json:"pwls"`
				LinkFlairCSSClass          interface{}   `json:"link_flair_css_class"`
				Downs                      int           `json:"downs"`
				ThumbnailHeight            int           `json:"thumbnail_height"`
				ParentWhitelistStatus      string        `json:"parent_whitelist_status"`
				HideScore                  bool          `json:"hide_score"`
				Name                       string        `json:"name"`
				Quarantine                 bool          `json:"quarantine"`
				LinkFlairTextColor         string        `json:"link_flair_text_color"`
				UpvoteRatio                float64       `json:"upvote_ratio"`
				AuthorFlairBackgroundColor interface{}   `json:"author_flair_background_color"`
				SubredditType              string        `json:"subreddit_type"`
				Ups                        int           `json:"ups"`
				Domain                     string        `json:"domain"`
				MediaEmbed                 struct {
				} `json:"media_embed"`
				ThumbnailWidth        int         `json:"thumbnail_width"`
				AuthorFlairTemplateID interface{} `json:"author_flair_template_id"`
				IsOriginalContent     bool        `json:"is_original_content"`
				AuthorFullname        string      `json:"author_fullname"`
				SecureMedia           interface{} `json:"secure_media"`
				IsRedditMediaDomain   bool        `json:"is_reddit_media_domain"`
				IsMeta                bool        `json:"is_meta"`
				Category              interface{} `json:"category"`
				SecureMediaEmbed      struct {
				} `json:"secure_media_embed"`
				LinkFlairText       interface{}   `json:"link_flair_text"`
				CanModPost          bool          `json:"can_mod_post"`
				Score               int           `json:"score"`
				ApprovedBy          interface{}   `json:"approved_by"`
				Thumbnail           string        `json:"thumbnail"`
				Edited              bool          `json:"edited"`
				AuthorFlairCSSClass interface{}   `json:"author_flair_css_class"`
				AuthorFlairRichtext []interface{} `json:"author_flair_richtext"`
				Gildings            struct {
					Gid1 int `json:"gid_1"`
					Gid2 int `json:"gid_2"`
					Gid3 int `json:"gid_3"`
				} `json:"gildings"`
				PostHint          string      `json:"post_hint"`
				ContentCategories interface{} `json:"content_categories"`
				IsSelf            bool        `json:"is_self"`
				ModNote           interface{} `json:"mod_note"`
				Created           float64     `json:"created"`
				LinkFlairType     string      `json:"link_flair_type"`
				Wls               int         `json:"wls"`
				BannedBy          interface{} `json:"banned_by"`
				AuthorFlairType   string      `json:"author_flair_type"`
				ContestMode       bool        `json:"contest_mode"`
				SelftextHTML      interface{} `json:"selftext_html"`
				Likes             interface{} `json:"likes"`
				SuggestedSort     interface{} `json:"suggested_sort"`
				BannedAtUtc       interface{} `json:"banned_at_utc"`
				ViewCount         interface{} `json:"view_count"`
				Archived          bool        `json:"archived"`
				NoFollow          bool        `json:"no_follow"`
				IsCrosspostable   bool        `json:"is_crosspostable"`
				Pinned            bool        `json:"pinned"`
				Over18            bool        `json:"over_18"`
				Preview           struct {
					Images []struct {
						Source struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"source"`
						Resolutions []struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"resolutions"`
						Variants struct {
							Obfuscated struct {
								Source struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"source"`
								Resolutions []struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"resolutions"`
							} `json:"obfuscated"`
							Nsfw struct {
								Source struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"source"`
								Resolutions []struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"resolutions"`
							} `json:"nsfw"`
						} `json:"variants"`
						ID string `json:"id"`
					} `json:"images"`
					Enabled bool `json:"enabled"`
				} `json:"preview"`
				Media                    interface{}   `json:"media"`
				MediaOnly                bool          `json:"media_only"`
				LinkFlairTemplateID      interface{}   `json:"link_flair_template_id"`
				CanGild                  bool          `json:"can_gild"`
				Spoiler                  bool          `json:"spoiler"`
				Locked                   bool          `json:"locked"`
				AuthorFlairText          interface{}   `json:"author_flair_text"`
				Visited                  bool          `json:"visited"`
				NumReports               interface{}   `json:"num_reports"`
				Distinguished            interface{}   `json:"distinguished"`
				SubredditID              string        `json:"subreddit_id"`
				ModReasonBy              interface{}   `json:"mod_reason_by"`
				RemovalReason            interface{}   `json:"removal_reason"`
				LinkFlairBackgroundColor string        `json:"link_flair_background_color"`
				ID                       string        `json:"id"`
				IsRobotIndexable         bool          `json:"is_robot_indexable"`
				ReportReasons            interface{}   `json:"report_reasons"`
				Author                   string        `json:"author"`
				NumCrossposts            int           `json:"num_crossposts"`
				NumComments              int           `json:"num_comments"`
				SendReplies              bool          `json:"send_replies"`
				AuthorPatreonFlair       bool          `json:"author_patreon_flair"`
				AuthorFlairTextColor     interface{}   `json:"author_flair_text_color"`
				Permalink                string        `json:"permalink"`
				WhitelistStatus          string        `json:"whitelist_status"`
				Stickied                 bool          `json:"stickied"`
				URL                      string        `json:"url"`
				SubredditSubscribers     int           `json:"subreddit_subscribers"`
				CreatedUtc               float64       `json:"created_utc"`
				ModReports               []interface{} `json:"mod_reports"`
				IsVideo                  bool          `json:"is_video"`
			} `json:"data"`
		} `json:"children"`
		After  interface{} `json:"after"`
		Before interface{} `json:"before"`
	} `json:"data"`
}
```
