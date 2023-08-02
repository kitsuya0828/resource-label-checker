package label

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// The title of the zip file displayed on Slack
	ZipFileTitle = "リソースタイプ別リソース名一覧"
)

// GetMainMessageText gets the text of the main message
func GetMainMessageText(result *Result, projectId string, date string) string {
	message := fmt.Sprintf("%s の *%s* の結果 :mag:\n", date, projectId)

	// Only resources without required labels are reported at this time
	sum := 0
	for _, v := range result.NoRequiredLabelResources {
		sum += len(v)
	}
	message += fmt.Sprintf(":label: 必須ラベルが付与されていないリソースは *%d* 件でした", sum)

	// You can change the message as you like
	return message
}

// GetThreadMessageText gets the text of the message in the thread
func GetThreadMessageText(result *Result) string {
	// Only resources without required labels are reported at this time
	message := "必須ラベルが付与されていないリソース一覧\n"
	keys := make([]string, 0, len(result.NoRequiredLabelResources))
	for k := range result.NoRequiredLabelResources {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		message += fmt.Sprintf("\n:pushpin: %s \n", k)
		resources := result.NoRequiredLabelResources[k]
		message += strings.Join(resources, "\n")
	}

	// You can change the message as you like
	return message
}
