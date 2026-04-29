package sloggin

import (
	"regexp"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

type Filter func(ctx *gin.Context) bool

// Basic
func Accept(filter Filter) Filter { return filter }
func Ignore(filter Filter) Filter { return func(ctx *gin.Context) bool { return !filter(ctx) } }

// Method
func AcceptMethod(methods ...string) Filter {
	for i := range methods {
		methods[i] = strings.ToLower(methods[i])
	}

	return func(c *gin.Context) bool {
		return slices.Contains(methods, strings.ToLower(c.Request.Method))
	}
}

func IgnoreMethod(methods ...string) Filter {
	for i := range methods {
		methods[i] = strings.ToLower(methods[i])
	}

	return func(c *gin.Context) bool {
		return !slices.Contains(methods, strings.ToLower(c.Request.Method))
	}
}

// Status
func AcceptStatus(statuses ...int) Filter {
	return func(c *gin.Context) bool {
		return slices.Contains(statuses, c.Writer.Status())
	}
}

func IgnoreStatus(statuses ...int) Filter {
	return func(c *gin.Context) bool {
		return !slices.Contains(statuses, c.Writer.Status())
	}
}

func AcceptStatusGreaterThan(status int) Filter {
	return func(c *gin.Context) bool {
		return c.Writer.Status() > status
	}
}

func AcceptStatusGreaterThanOrEqual(status int) Filter {
	return func(c *gin.Context) bool {
		return c.Writer.Status() >= status
	}
}

func AcceptStatusLessThan(status int) Filter {
	return func(c *gin.Context) bool {
		return c.Writer.Status() < status
	}
}

func AcceptStatusLessThanOrEqual(status int) Filter {
	return func(c *gin.Context) bool {
		return c.Writer.Status() <= status
	}
}

func IgnoreStatusGreaterThan(status int) Filter {
	return AcceptStatusLessThanOrEqual(status)
}

func IgnoreStatusGreaterThanOrEqual(status int) Filter {
	return AcceptStatusLessThan(status)
}

func IgnoreStatusLessThan(status int) Filter {
	return AcceptStatusGreaterThanOrEqual(status)
}

func IgnoreStatusLessThanOrEqual(status int) Filter {
	return AcceptStatusGreaterThan(status)
}

// Path
func AcceptPath(urls ...string) Filter {
	return func(c *gin.Context) bool {
		return slices.Contains(urls, c.Request.URL.Path)
	}
}

func IgnorePath(urls ...string) Filter {
	return func(c *gin.Context) bool {
		return !slices.Contains(urls, c.Request.URL.Path)
	}
}

func AcceptPathContains(parts ...string) Filter {
	return func(c *gin.Context) bool {
		for _, part := range parts {
			if strings.Contains(c.Request.URL.Path, part) {
				return true
			}
		}

		return false
	}
}

func IgnorePathContains(parts ...string) Filter {
	return func(c *gin.Context) bool {
		for _, part := range parts {
			if strings.Contains(c.Request.URL.Path, part) {
				return false
			}
		}

		return true
	}
}

func AcceptPathPrefix(prefixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(c.Request.URL.Path, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnorePathPrefix(prefixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(c.Request.URL.Path, prefix) {
				return false
			}
		}

		return true
	}
}

func AcceptPathSuffix(suffixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(c.Request.URL.Path, suffix) {
				return true
			}
		}

		return false
	}
}

func IgnorePathSuffix(suffixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(c.Request.URL.Path, suffix) {
				return false
			}
		}

		return true
	}
}

func AcceptPathMatch(regs ...regexp.Regexp) Filter {
	return func(c *gin.Context) bool {
		for _, reg := range regs {
			if reg.MatchString(c.Request.URL.Path) {
				return true
			}
		}

		return false
	}
}

func IgnorePathMatch(regs ...regexp.Regexp) Filter {
	return func(c *gin.Context) bool {
		for _, reg := range regs {
			if reg.MatchString(c.Request.URL.Path) {
				return false
			}
		}

		return true
	}
}

// Host
func AcceptHost(hosts ...string) Filter {
	return func(c *gin.Context) bool {
		return slices.Contains(hosts, c.Request.URL.Host)
	}
}

func IgnoreHost(hosts ...string) Filter {
	return func(c *gin.Context) bool {
		return !slices.Contains(hosts, c.Request.URL.Host)
	}
}

func AcceptHostContains(parts ...string) Filter {
	return func(c *gin.Context) bool {
		for _, part := range parts {
			if strings.Contains(c.Request.URL.Host, part) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostContains(parts ...string) Filter {
	return func(c *gin.Context) bool {
		for _, part := range parts {
			if strings.Contains(c.Request.URL.Host, part) {
				return false
			}
		}

		return true
	}
}

func AcceptHostPrefix(prefixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(c.Request.URL.Host, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostPrefix(prefixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(c.Request.URL.Host, prefix) {
				return false
			}
		}

		return true
	}
}

func AcceptHostSuffix(suffixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(c.Request.URL.Host, suffix) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostSuffix(suffixs ...string) Filter {
	return func(c *gin.Context) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(c.Request.URL.Host, suffix) {
				return false
			}
		}

		return true
	}
}

func AcceptHostMatch(regs ...regexp.Regexp) Filter {
	return func(c *gin.Context) bool {
		for _, reg := range regs {
			if reg.MatchString(c.Request.URL.Host) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostMatch(regs ...regexp.Regexp) Filter {
	return func(c *gin.Context) bool {
		for _, reg := range regs {
			if reg.MatchString(c.Request.URL.Host) {
				return false
			}
		}

		return true
	}
}
