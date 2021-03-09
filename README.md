# Panggil HTTP
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/KodepandaID/panggilhttp)
![GitHub](https://img.shields.io/github/license/KodepandaID/panggilhttp)
![](https://github.com/KodepandaID/panggilhttp/workflows/Go/badge.svg)

An enhanced HTTP client for Go with features likes:
- Support call GET Method more than 1 URL and merged the response body.
- Set which values from the response body to show with Withelist or Blacklist.
- HTTP retry if failed, with attempts and interval configuration.

## Installation
```bash
go get github.com/KodepandaID/panggilhttp
```

## Example
```go
import "github.com/KodepandaID/panggilhttp"

func main() {
    client := panggilhttp.New()
    
    resp, e := client.
		Get("http://localhost:3000/hotels", nil, nil).
		Do()
	if e != nil {
		panic(e)
	}
}
```

## License

Copyright [Yudha Pratama Wicaksana](https://github.com/LordAur), Licensed under [MIT](./LICENSE).