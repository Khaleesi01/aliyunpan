// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package command

import (
	"fmt"
	"github.com/tickstep/aliyunpan-api/aliyunpan"
	"github.com/tickstep/aliyunpan/internal/config"
	"github.com/tickstep/library-go/logger"
	"math/rand"
	"net/url"
	"path"
	"strings"
	"time"
)

var (
	panCommandVerbose = logger.New("PANCOMMAND", config.EnvVerbose)
)

const(
	// 备份数据库桶分区标志
	BackupMetaBucketName = "adrive"

	// 备份数据文件夹目录名称，隐藏目录
	BackupMetaDirName = ".adrive"
)

// GetFileInfoByPaths 获取指定文件路径的文件详情信息
func GetAppFileInfoByPaths(driveId string, paths ...string) (fileInfoList []*aliyunpan.FileEntity, failedPaths []string, error error) {
	if len(paths) <= 0 {
		return nil, nil, fmt.Errorf("请指定文件路径")
	}
	activeUser := GetActiveUser()

	for idx := 0; idx < len(paths); idx++ {
		absolutePath := path.Clean(activeUser.PathJoin(driveId, paths[idx]))
		fe, err := activeUser.PanClient().FileInfoByPath(driveId, absolutePath)
		if err != nil {
			failedPaths = append(failedPaths, absolutePath)
			continue
		}
		fileInfoList = append(fileInfoList, fe)
	}
	return
}

// GetFileInfoByPaths 获取指定文件路径的文件详情信息
func GetFileInfoByPaths(paths ...string) (fileInfoList []*aliyunpan.FileEntity, failedPaths []string, error error) {
	if len(paths) <= 0 {
		return nil, nil, fmt.Errorf("请指定文件路径")
	}
	activeUser := GetActiveUser()

	for idx := 0; idx < len(paths); idx++ {
		absolutePath := path.Clean(activeUser.PathJoin(activeUser.ActiveDriveId, paths[idx]))
		fe, err := activeUser.PanClient().FileInfoByPath(activeUser.ActiveDriveId, absolutePath)
		if err != nil {
			failedPaths = append(failedPaths, absolutePath)
			continue
		}
		fileInfoList = append(fileInfoList, fe)
	}
	return
}

func matchPathByShellPattern(driveId string, patterns ...string) (panpaths []string, err error) {
	acUser := GetActiveUser()
	for k := range patterns {
		ps := acUser.PathJoin(driveId, patterns[k])
		panpaths = append(panpaths, ps)
	}
	return panpaths, nil
}

func RandomStr(count int) string {
	//STR_SET := "abcdefjhijklmnopqrstuvwxyzABCDEFJHIJKLMNOPQRSTUVWXYZ1234567890"
	STR_SET := "abcdefjhijklmnopqrstuvwxyz1234567890"
	rand.Seed(time.Now().UnixNano())
	str := strings.Builder{}
	for i := 0; i < count; i++  {
		str.WriteByte(byte(STR_SET[rand.Intn(len(STR_SET))]))
	}
	return str.String()
}

func GetAllPathFolderByPath(pathStr string) []string {
	dirNames := strings.Split(pathStr, "/")
	dirs := []string{}
	p := "/"
	dirs = append(dirs, p)
	for _,s := range dirNames {
		p = path.Join(p, s)
		dirs = append(dirs, p)
	}
	return dirs
}

// EscapeStr 转义字符串
func EscapeStr(s string) string {
	return url.PathEscape(s)
}

// UnescapeStr 反转义字符串
func UnescapeStr(s string) string {
	r,_ := url.PathUnescape(s)
	return r
}

// RefreshTokenInNeed 刷新refresh token
func RefreshTokenInNeed(activeUser *config.PanUser) bool {
	if activeUser == nil {
		return false
	}

	// refresh expired token
	if activeUser.PanClient() != nil {
		if len(activeUser.WebToken.RefreshToken) > 0 {
			cz := time.FixedZone("CST", 8*3600) // 东8区
			expiredTime, _ := time.ParseInLocation("2006-01-02 15:04:05", activeUser.WebToken.ExpireTime, cz)
			now := time.Now()
			if (expiredTime.Unix() - now.Unix()) <= (10 * 60) { // 10min
				// need update refresh token
				logger.Verboseln("access token expired, get new from refresh token")
				if wt, er := aliyunpan.GetAccessTokenFromRefreshToken(activeUser.RefreshToken); er == nil {
					activeUser.WebToken = *wt
					activeUser.PanClient().UpdateToken(*wt)
					logger.Verboseln("get new access token success")
					return true
				}
			}
		}
	}
	return false
}