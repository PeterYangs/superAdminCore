package route

import (
	"github.com/PeterYangs/superAdminCore/v2/contextPlus"
	"github.com/PeterYangs/superAdminCore/v2/kernel"
	"github.com/PeterYangs/superAdminCore/v2/response"
	"github.com/PeterYangs/superAdminCore/v2/route/allUrl"
	"github.com/gin-gonic/gin"
	"regexp"
)

const (
	GET    int = 0x000000
	POST   int = 0x000001
	PUT    int = 0x000002
	DELETE int = 0x000003
	ANY    int = 0x000004
)

type router struct {
	engine *gin.Engine
}

type Group struct {
	engine      *gin.Engine
	middlewares []contextPlus.HandlerFunc
	path        string
}

type handler struct {
	handlerFunc func(*contextPlus.Context) *response.Response
	middlewares []contextPlus.HandlerFunc
	engine      *gin.Engine
	url         string
	method      int
	regex       map[string]string //路由正则表达式
	Tag         string            //函数标记
}

func newRouter(engine *gin.Engine) *router {

	return &router{
		engine: engine,
	}
}

func (rr *router) Group(path string, callback func(Group), middlewares ...contextPlus.HandlerFunc) {

	g := Group{
		engine:      rr.engine,
		middlewares: middlewares,
		path:        dealPath(path),
	}

	callback(g)

}

func (gg Group) Group(path string, callback func(group2 Group), middlewares ...contextPlus.HandlerFunc) {

	gg.middlewares = append(gg.middlewares, middlewares...)

	gg.path += dealPath(path)

	callback(gg)

}

func (gg Group) Registered(method int, url string, f func(c *contextPlus.Context) *response.Response, middlewares ...contextPlus.HandlerFunc) *handler {

	gg.middlewares = append(gg.middlewares, middlewares...)

	return &handler{
		handlerFunc: f,
		engine:      gg.engine,
		url:         gg.path + dealPath(url),
		method:      method,
		middlewares: gg.middlewares,
	}

}

func (h *handler) Regex(reg map[string]string) *handler {

	h.regex = reg

	return h
}

func (h *handler) SetTag(tag string) *handler {

	h.Tag = tag

	return h
}

func (h *handler) Bind() {

	ff := func(c *contextPlus.Context) {

		data := h.handlerFunc(c)

		getDataType(data.GetData(), c)

	}

	//控制器放最前面
	h.middlewares = append(h.middlewares, ff)

	var temp = make([]gin.HandlerFunc, len(h.middlewares))

	for i, funcs := range h.middlewares {

		tempFuncs := funcs

		temp[i] = func(context *gin.Context) {

			tempFuncs(&contextPlus.Context{
				Context: context,
				Handler: &contextPlus.Handler{
					HandlerFunc: h.handlerFunc,
					Engine:      h.engine,
					Url:         h.url,
					Method:      h.method,
					Regex:       h.regex,
					Tag:         h.Tag,
				},
			})

		}

	}

	all := allUrl.NewAllUrl()

	all.Add(h.url)

	switch h.method {

	case GET:

		h.engine.GET(h.url, temp...)

	case POST:

		h.engine.POST(h.url, temp...)

	case PUT:

		h.engine.PUT(h.url, temp...)

	case DELETE:

		h.engine.DELETE(h.url, temp...)

	case ANY:

		h.engine.Any(h.url, temp...)
	}

}

func getDataType(data interface{}, c *contextPlus.Context) {

	switch item := data.(type) {

	case map[string]interface{}:

		c.JSON(200, item)

	case string:

		c.String(200, item)

	case []byte:

		c.String(200, string(item))
	case gin.H:

		c.JSON(200, item)
	case nil:

	}
}

func Load(rr *gin.Engine, routes func(g Group)) {

	//fmt.Println("xxxxxxxxxxxxxxxxxxxxxx")

	_r := newRouter(rr)

	_r.Group("", func(global Group) {

		//判断http服务已启动接口
		global.Registered(GET, "/_ping/:id", func(c *contextPlus.Context) *response.Response {

			id := c.Param("id")

			//判断服务id
			if id == kernel.Id {

				return response.Resp().String("success")
			}

			return response.Resp().String("fail")

		}).Bind()

		routes(global)

	}, kernel.Middleware...)

}

func dealPath(path string) string {

	re1, err := regexp.MatchString("^/.*", path)

	if err != nil {

		return path
	}

	if re1 {

		return path
	}

	return "/" + path
}
