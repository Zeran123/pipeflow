package bot

import (
	"strings"

	"github.com/buger/jsonparser"
)

type GitlabEvent struct {
	ProjectName        string
	ProjectUrl         string
	RepositoryName     string
	RepositoryHomePage string
	UserName           string
}

type GitlabMergeRequestEvent struct {
	GitlabEvent
	MergeRequestId int64
	Title          string
	Url            string
	SourceBranch   string
	TargetBranch   string
	Status         string
	WorkInProgress bool
	Assignee       GitlabAssignee
}

type GitlabPushEvent struct {
	GitlabEvent
	Branch  string
	Commits []GitlabCommit
}

type GitlabCommit struct {
	Id      string
	Message string
	Url     string
}

type GitlabAssignee struct {
	AssigneeName string
}

type GitlabComment struct {
	GitlabEvent
	Note         string
	NoteType     string
	Url          string
	MergeRequest GitlabCommentOnMergeRequest
	Commit       GitlabCommit
	Issue        GitlabIssue
}

type GitlabCommentOnMergeRequest struct {
	Id    int64
	Title string
}

type GitlabIssue struct {
	Id    int64
	Title string
}

func (e GitlabMergeRequestEvent) Format() string {
	return Format(e, "tmpl/gitlab/mergerequest2wechat.tmpl")
}

func (e GitlabPushEvent) Format() string {
	return Format(e, "tmpl/gitlab/push2wechat.tmpl")
}

func (e GitlabComment) Format() string {
	return Format(e, "tmpl/gitlab/comment2wechat.tmpl")
}

func ProcessFromGitlab(s Store, rawData []byte) {
	var events []interface{}
	o := JsonObj{rawData}
	event := o.GetStr("object_kind")
	if s.Target == wechatTarget {
		if event == "merge_request" {
			events = readMergeRequestContentFromGitlab(rawData)
		} else if event == "push" {
			events = readPushContentFromGitlab(rawData)
		} else if event == "note" {
			events = readCommentContentFromGitlab(rawData)
		}
	}
	Send2Wechat(s, events)
}

func readMergeRequestContentFromGitlab(rawData []byte) []interface{} {
	gitlabs := make([]interface{}, 0, 1)
	o := JsonObj{rawData}
	msg := GitlabMergeRequestEvent{}
	msg.MergeRequestId = o.GetInt("object_attributes", "iid")
	msg.Title = o.GetStr("object_attributes", "title")
	msg.Url = o.GetStr("object_attributes", "url")
	msg.Status = o.GetStr("object_attributes", "state")
	msg.UserName = o.GetStr("user", "name")
	msg.RepositoryName = o.GetStr("repository", "name")
	msg.RepositoryHomePage = o.GetStr("repository", "homepage")
	msg.ProjectName = o.GetStr("project", "name")
	msg.Assignee = GitlabAssignee{o.GetStr("object_attributes", "assignee", "name")}
	gitlabs = append(gitlabs, msg)
	return gitlabs
}

func readPushContentFromGitlab(rawData []byte) []interface{} {
	gitlabs := make([]interface{}, 0, 1)
	o := JsonObj{rawData}
	msg := GitlabPushEvent{}
	msg.UserName = o.GetStr("user_name")
	msg.RepositoryName = o.GetStr("repository", "name")
	msg.RepositoryHomePage = o.GetStr("repository", "homepage")
	msg.ProjectName = o.GetStr("project", "name")
	msg.Branch = o.GetStr("ref")
	msg.Branch = msg.Branch[strings.LastIndex(msg.Branch, "/")+1 : len(msg.Branch)]
	commits := make([]GitlabCommit, 0, 5)
	jsonparser.ArrayEach(rawData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		co := JsonObj{value}
		c := GitlabCommit{}
		c.Id = strings.ToUpper(co.GetStr("id")[0:6])
		c.Message = co.GetStr("message")
		commits = append(commits, c)
	}, "commits")
	msg.Commits = commits
	gitlabs = append(gitlabs, msg)
	return gitlabs
}

func readCommentContentFromGitlab(rawData []byte) []interface{} {
	gitlabs := make([]interface{}, 0, 1)
	o := JsonObj{rawData}
	msg := GitlabComment{}
	msg.Note = o.GetStr("object_attributes", "note")
	msg.NoteType = o.GetStr("object_attributes", "noteable_type")
	msg.UserName = o.GetStr("user", "name")
	msg.RepositoryName = o.GetStr("repository", "name")
	msg.RepositoryHomePage = o.GetStr("repository", "homepage")
	msg.ProjectName = o.GetStr("project", "name")
	msg.Url = o.GetStr("object_attributes", "url")
	msg.MergeRequest = GitlabCommentOnMergeRequest{o.GetInt("merge_request", "iid"), o.GetStr("merge_request", "title")}
	msg.Commit = GitlabCommit{o.GetStr("commit", "id"), o.GetStr("commit", "message"), o.GetStr("commit", "url")}
	if msg.Commit.Id != "" {
		msg.Commit.Id = strings.ToUpper(msg.Commit.Id[0:6])
	}
	msg.Issue = GitlabIssue{o.GetInt("issue", "iid"), o.GetStr("issue", "title")}
	gitlabs = append(gitlabs, msg)
	return gitlabs
}
