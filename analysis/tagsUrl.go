package analysis

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type tagsUrlType map[string][]string
type rsType struct {
	list  []string
	count int
}

var (
	tagsUrlArr   = tagsUrlType{}
	tagsRsUrlArr = tagsUrlType{}
	tagTempDir   = "./__au_tag_temp__"
	tagRsPath    string
	tagRelReg    = regexp.MustCompile(`["|']`)
)

func TagsUrl(filePath string, cwd string, fileName string) {
	tagsUrlArr = tagsUrlType{}
	tagsRsUrlArr = tagsUrlType{}
	fi, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		tmpStr := string(a)
		if !tagRelReg.MatchString(tmpStr) && tagRe.MatchString(tmpStr) {
			getTags(tmpStr)
		}
	}

	var buf bytes.Buffer
	for k, v := range tagsUrlArr {
		buf.WriteString(k)
		buf.WriteString(" ")
		l := len(v)
		if l > 10 {
			l = 10
		}
		buf.WriteString(strings.Join(v[:l], ","))
		buf.WriteString("\n")
	}

	tagRsPath = ensureDir(filepath.Join(cwd, tagTempDir))
	finalPath := filepath.Join(tagRsPath, fileName+tempExt)
	fmt.Printf("\nMerge file in %v\n", finalPath)
	if e := ioutil.WriteFile(finalPath, []byte(buf.String()), 0777); e != nil {
		log.Fatal(e)
	}
}

func getTags(c string) {
	tagsInfo := pluginRe.FindAllStringSubmatch(c, -1)
	if len(tagsInfo) > 0 && len(tagsInfo[0]) > 1 {
		url := tagsInfo[0][1]
		tags := tagsInfo[0][2]
		tagsArr := strings.Split(tags, ", ")

		for _, v := range tagsArr {
			tagsUrlArr[v] = append(tagsUrlArr[v], url)
		}
	}
}

func GetTagsMap(cwd string, anaDate string) {
	files, err := ioutil.ReadDir(tagRsPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fi, err := os.Open(filepath.Join(tagRsPath, file.Name()))
		if err != nil {
			log.Fatal(err)
			return
		}
		defer fi.Close()
		br := bufio.NewReader(fi)
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			content := string(a)
			infos := strings.Split(content, " ")
			if len(infos) > 1 {
				tag := infos[0]
				urlArr := strings.Split(infos[1], ",")

				if len(tagsRsUrlArr[tag]) > 0 {
					tagsRsUrlArr[tag] = append(tagsRsUrlArr[tag], urlArr...)
				} else {
					tagsRsUrlArr[tag] = urlArr
				}
			}
		}
	}

	for k, v := range tagsRsUrlArr {
		sort.Strings(v)
		tagsRsUrlArr[k] = uniq(v)
	}
	bArr := []string{}
	for k, v := range tagsRsUrlArr {
		rl := len(v)
		if rl > 10 {
			rl = 10
		}
		tmp := strings.Join(v[:rl], ",")
		bArr = append(bArr, "('"+k+"', '"+tmp+"', '0', '"+anaDate+"', '"+time.Now().String()+"')")
	}
	openDb(cwd)
	sqlStr := "INSERT INTO tags (tag_name, urls, url_count, ana_date, edit_date) VALUES " + strings.Join(bArr, ",")
	rs, err := db.Exec(sqlStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rs)

	defer db.Close()
}
