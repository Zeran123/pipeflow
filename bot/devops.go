package bot

import (
	"regexp"
	"strings"
)

type DevOpsCreateWorkItem struct {
	DevOpsWorkItem
	CreatedBy string
}

type DevOpsUpdateWorkItem struct {
	DevOpsWorkItem
	UpdatedBy string
}

type DevOpsCommentWorkItem struct {
	DevOpsWorkItem
	Comment     string
	ActivatedBy string
}

type DevOpsWorkItem struct {
	Resource     DevOpsResource
	WorkItemType string
	ProjectName  string
	AssignedTo   string
	State        string
	Reason       string
}

type DevOpsResource struct {
	Id    int64
	Title string
	Url   string
}

func (i DevOpsCreateWorkItem) Format() string {
	return Format(i, "tmpl/devops/createworkitem2wechat.tmpl")
}

func (i DevOpsUpdateWorkItem) Format() string {
	return Format(i, "tmpl/devops/updateworkitem2wechat.tmpl")
}

func (i DevOpsCommentWorkItem) Format() string {
	return Format(i, "tmpl/devops/commentworkitem2wechat.tmpl")
}

func ProcessFromDevOps(s Store, rawData []byte) {
	var events []interface{}
	o := JsonObj{rawData}
	event := o.GetStr("eventType")
	if s.Target == wechatTarget {
		if event == "workitem.created" {
			events = readWorkItemCreatedContenxtFromDevOps(rawData)
		} else if event == "workitem.updated" {
			events = readWorkItemUpdateContenxtFromDevOps(rawData)
		} else if event == "workitem.commented" {
			events = readWorkItemCommentContenxtFromDevOps(rawData)
		}
	}
	Send2Wechat(s, events)
}

func readWorkItemCreatedContenxtFromDevOps(rawData []byte) []interface{} {
	contents := make([]interface{}, 0, 1)
	o := JsonObj{rawData}
	item := DevOpsCreateWorkItem{}
	item.WorkItemType = o.GetStr("resource", "fields", "System.WorkItemType")
	item.CreatedBy = getUserName(o.GetStr("resource", "fields", "System.CreatedBy"))
	item.ProjectName = o.GetStr("resource", "fields", "System.TeamProject")
	item.AssignedTo = getUserName(o.GetStr("resource", "fields", "System.AssignedTo"))
	item.State = o.GetStr("resource", "fields", "System.State")
	item.Resource.Id = o.GetInt("resource", "id")
	item.Resource.Title = o.GetStr("resource", "fields", "System.Title")
	item.Resource.Url = o.GetStr("resource", "_links", "html", "href")
	contents = append(contents, item)
	return contents
}

func readWorkItemUpdateContenxtFromDevOps(rawData []byte) []interface{} {
	contents := make([]interface{}, 0, 1)
	o := JsonObj{rawData}
	item := DevOpsUpdateWorkItem{}
	item.WorkItemType = o.GetStr("resource", "revision", "fields", "System.WorkItemType")
	item.UpdatedBy = getUserName(o.GetStr("resource", "revision", "fields", "System.ChangedBy"))
	item.ProjectName = o.GetStr("resource", "revision", "fields", "System.TeamProject")
	item.AssignedTo = getUserName(getNewValueIfExists(o, "System.AssignedTo"))
	item.State = getNewValueIfExists(o, "System.State")
	item.Reason = getNewValueIfExists(o, "System.Reason")
	item.Resource.Id = o.GetInt("resource", "id")
	item.Resource.Title = getNewValueIfExists(o, "System.Title")
	item.Resource.Url = o.GetStr("resource", "_links", "html", "href")
	contents = append(contents, item)
	return contents
}

func readWorkItemCommentContenxtFromDevOps(rawData []byte) []interface{} {
	contents := make([]interface{}, 0, 1)
	o := JsonObj{rawData}
	item := DevOpsCommentWorkItem{}
	item.WorkItemType = o.GetStr("resource", "fields", "System.WorkItemType")
	item.ActivatedBy = getUserName(o.GetStr("resource", "fields", "Microsoft.VSTS.Common.ActivatedBy"))
	item.ProjectName = o.GetStr("resource", "fields", "System.TeamProject")
	item.Comment = strings.Split(o.GetStr("detailedMessage", "text"), "\r\n")[3]
	matched, _ := regexp.MatchString("\\(mailto:.*\\)", item.Comment)
	if matched {
		re := regexp.MustCompile("\\(mailto:.*\\)")
		item.Comment = re.ReplaceAllString(item.Comment, "")
	}
	item.Resource.Id = o.GetInt("resource", "id")
	item.Resource.Title = o.GetStr("resource", "fields", "System.Title")
	item.Resource.Url = o.GetStr("resource", "_links", "html", "href")
	contents = append(contents, item)
	return contents
}

func getNewValueIfExists(o JsonObj, field string) string {
	newVal := o.GetStr("resource", "fields", field, "newValue")
	if newVal == "" {
		newVal = o.GetStr("resource", "revision", "fields", field)
	}
	return newVal
}

func getUserName(origin string) string {
	return strings.Split(origin, " ")[0]
}
