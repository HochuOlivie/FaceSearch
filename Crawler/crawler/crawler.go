package crawler

type ImageWithMetadata struct {
	ImageURL, PageURL string
	Title             string
}

type HttpMethod string

const (
	GET  HttpMethod = "GET"
	POST            = "POST"
)

type Request struct {
	Method      HttpMethod
	Url         string
	StrategyId  string
	ContentType string
	Body        string
}

type Strategy interface {
	Crawl(req Request) (err error, results []ImageWithMetadata, nextRequests []Request)
	Id() string
}
