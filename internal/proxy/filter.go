package proxy

import (
	"regexp"

	"peage/internal/config"
)

var (
	dockerPathPatterns       []*regexp.Regexp
	podmanPathPatterns       []*regexp.Regexp
	podmanCompatPathPatterns []*regexp.Regexp
	dockerVersionPattern   *regexp.Regexp
	podmanVersionPattern   *regexp.Regexp
)

func init() {
	dockerPathPatterns = compilePatterns([]string{
		"^/containers/json$",
		"^/containers/[^/]+/(json|stats)$",
		"^/events$",
		"^/images/json$",
		"^/images/[^/]+/json$",
		"^/info$",
		"^/networks$",
		"^/version$",
		"^/volumes$",
		"^/volumes/[^/]+$",
		"^/_ping$",
	})
	podmanPathPatterns = compilePatterns([]string{
		"^/libpod/containers/json$",
		"^/libpod/containers/stats$",
		"^/libpod/containers/[^/]+/(json|changes|exists|stats)$",
		"^/libpod/events$",
		"^/libpod/images/json$",
		"^/libpod/images/[^/]+/(json|exists)$",
		"^/libpod/info$",
		"^/libpod/networks/json$",
		"^/libpod/networks/[^/]+/(json|exists)$",
		"^/libpod/pods/json$",
		"^/libpod/pods/stats$",
		"^/libpod/pods/[^/]+/(json|exists)$",
		"^/libpod/_ping$",
		"^/libpod/version$",
		"^/libpod/volumes/json$",
		"^/libpod/volumes/[^/]+/(json|exists)$",
	})
	podmanCompatPathPatterns = make([]*regexp.Regexp, 0, len(dockerPathPatterns)+len(podmanPathPatterns))
	podmanCompatPathPatterns = append(podmanCompatPathPatterns, dockerPathPatterns...)
	podmanCompatPathPatterns = append(podmanCompatPathPatterns, podmanPathPatterns...)

	dockerVersionPattern = regexp.MustCompile(`^/v\d+\.\d+`)
	podmanVersionPattern = regexp.MustCompile(`^/v\d+\.\d+(\.\d+)?`)
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		compiled[i] = regexp.MustCompile(p)
	}
	return compiled
}

func getEngineVersionPattern(engine config.Engine) *regexp.Regexp {
	switch engine {
	case config.EngineDocker:
		return dockerVersionPattern
	case config.EnginePodman, config.EnginePodmanCompat:
		return podmanVersionPattern
	default:
		return nil
	}
}

func getEngineAllowedPaths(engine config.Engine) []*regexp.Regexp {
	switch engine {
	case config.EngineDocker:
		return dockerPathPatterns
	case config.EnginePodman:
		return podmanPathPatterns
	case config.EnginePodmanCompat:
		return podmanCompatPathPatterns
	default:
		return nil
	}
}

func isAllowedPath(path string, engine config.Engine) bool {
	versionPattern := getEngineVersionPattern(engine)
	if versionPattern != nil {
		path = versionPattern.ReplaceAllString(path, "")
	}

	for _, p := range getEngineAllowedPaths(engine) {
		if p.MatchString(path) {
			return true
		}
	}
	return false
}
