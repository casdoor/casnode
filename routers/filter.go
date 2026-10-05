// Copyright 2020 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routers

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/context"
	"github.com/casbin/casnode/object"
	"github.com/casbin/casnode/util"
)

const defaultDescription = "Casnode is an open-source forum (BBS) software written in Go and React"

func Static(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path
	if strings.HasPrefix(urlPath, "/api/") {
		return
	}

	path := "web/build"
	if urlPath == "/" {
		path += "/index.html"
	} else {
		path += urlPath
	}

	if util.FolderExist(path) {
		ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		return
	}

	if util.FileExist(path) {
		if path == "web/build/index.html" {
			FreshAccountActiveStatus(ctx)
			serveIndexHtml(ctx)
			return
		}

		http.ServeFile(ctx.ResponseWriter, ctx.Request, path)
	} else {
		serveIndexHtml(ctx)
	}
}

func serveIndexHtml(ctx *context.Context) {
	content, err := os.ReadFile("web/build/index.html")
	if err != nil {
		panic(err)
	}

	res := string(content)
	if conf := object.GetFrontConfById("forumName"); conf != nil && conf.Value != "" {
		res = strings.Replace(res, "<title>Casnode</title>", fmt.Sprintf("<title>%s</title>", html.EscapeString(conf.Value)), 1)
	}
	if conf := object.GetFrontConfById("signinBoxSpan"); conf != nil && conf.Value != "" {
		res = strings.Replace(res, fmt.Sprintf(`content="%s"`, defaultDescription), fmt.Sprintf(`content="%s"`, html.EscapeString(conf.Value)), 1)
	}

	http.ServeContent(ctx.ResponseWriter, ctx.Request, "index.html", time.Time{}, strings.NewReader(res))
}

// FreshAccountActiveStatus fresh member's online status.
func FreshAccountActiveStatus(ctx *context.Context) {
	claims := getSessionClaims(ctx)
	if claims == nil {
		return
	}

	_, err := object.UpdateMemberOnlineStatus(&claims.User, true, util.GetCurrentTime())
	if err != nil {
		panic(err)
	}
}
