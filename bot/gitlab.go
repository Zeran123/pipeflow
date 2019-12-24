package bot

import (
	"bytes"
	"text/template"
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

type GitlabAssignee struct {
	AssigneeName string
}

func (e GitlabMergeRequestEvent) Format() string {
	buf := new(bytes.Buffer)
	tmpl, _ := template.ParseFiles("tmpl/gitlab/mergerequest2wechat.tmpl")
	tmpl.Execute(buf, e)
	return buf.String()
}

func ProcessFromGitlab(s Store, rawData []byte) {
	o := JsonObj{rawData}
	event := o.GetStr("object_kind")
	if s.Target == wechatTarget {
		if event == "merge_request" {
			events := readMergeRequestContentFromGitlab(rawData)
			Send2Wechat(s, events)
		}
	}
}

func readMergeRequestContentFromGitlab(rawData []byte) []interface{} {
	gitlabs := make([]interface{}, 0, 5)
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
