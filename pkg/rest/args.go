/*
Copyright 2019 Alexander Sowitzki.

GNU Affero General Public License version 3 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://opensource.org/licenses/AGPL-3.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Argument is an argument that can be collected from an URL.
type Argument func(url.URL) error

// Collect all the given arguments.
func CollectArguments(url url.URL, arguments []Argument) error {
	for _, argument := range arguments {
		if err := argument(url); err != nil {
			return err
		}
	}
	return nil
}

func stringFromURL(name string, url url.URL, set *bool) (string, error) {
	values, ok := url.Query()[name]

	if set != nil {
		*set = ok
	}

	switch {
	case ok && len(values) > 1:
		return "", fmt.Errorf("url contains %v multiple times", name)
	case ok:
		return values[0], nil
	case !ok && set != nil:
		return "", nil
	default:
		return "", fmt.Errorf("url does not contain argument %v", name)
	}
}

// BoolArgument represents a bool type argument.
func BoolArgument(name string, target *bool, set *bool) Argument {
	return func(url url.URL) error {
		rawValue, err := stringFromURL(name, url, set)
		switch {
		case err != nil:
			return err
		case set != nil && !*set:
			return nil
		}

		value, err := strconv.ParseBool(rawValue)
		if err != nil {
			return fmt.Errorf("could not parse bool from URL argument %v: %v", name, err)
		}
		*target = value
		return nil
	}
}

// StrippedURL without query is returned.
func StrippedURL(target *string) Argument {
	return func(url url.URL) error {
		*target = url.Scheme + url.Host + url.Path
		return nil
	}
}

// IntArgument represents a integer argument.
func IntArgument(name string, target *int, set *bool) Argument {
	return func(url url.URL) error {
		rawValue, err := stringFromURL(name, url, set)
		switch {
		case err != nil:
			return err
		case set != nil && !*set:
			return nil
		}

		value, err := strconv.ParseInt(rawValue, 10, 0)
		if err != nil {
			return fmt.Errorf("could not parse int from URL argument %v: %v", name, err)
		}
		*target = int(value)
		return nil
	}
}

// FloatArgument represents a float64 argument.
func FloatArgument(name string, target *float64, set *bool) Argument {
	return func(url url.URL) error {
		rawValue, err := stringFromURL(name, url, set)
		switch {
		case err != nil:
			return err
		case set != nil && !*set:
			return nil
		}

		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return fmt.Errorf("could not parse float from URL argument %v: %v", name, err)
		}
		*target = value
		return nil
	}
}

// UintArgument represents an unsigned integer argument.
func UintArgument(name string, target *uint, set *bool) Argument {
	return func(url url.URL) error {
		rawValue, err := stringFromURL(name, url, set)
		switch {
		case err != nil:
			return err
		case set != nil && !*set:
			return nil
		}

		value, err := strconv.ParseUint(rawValue, 10, 0)
		if err != nil {
			return fmt.Errorf("could not parse int from URL argument %v: %v", name, err)
		}
		*target = uint(value)
		return nil
	}
}

// DurationArgument represents an time duration argument.
func DurationArgument(name string, target *time.Duration, set *bool) Argument {
	return func(url url.URL) error {
		rawValue, err := stringFromURL(name, url, set)
		switch {
		case err != nil:
			return err
		case set != nil && !*set:
			return nil
		}

		value, err := time.ParseDuration(rawValue)
		if err != nil {
			return fmt.Errorf("could not parse duration from URL argument %v: %v", name, err)
		}
		*target = value
		return nil
	}
}

// StringArgument represents a string argument.
func StringArgument(name string, target *string, set *bool) Argument {
	return func(url url.URL) error {
		value, err := stringFromURL(name, url, set)
		switch {
		case err != nil:
			return err
		case set != nil && !*set:
			return nil
		}

		*target = value
		return nil
	}
}
