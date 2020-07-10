package testsupport

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// BackendKey allows to retrieve a resource from the backend using its ID.
//
// BackendKey must return true if the resource was found.
type BackendKey func(c *hcloud.Client, id int) bool

// CheckResourceExists checks that a resource with the passed name exists.
//
// CheckResourceExists uses k to actually retrieve the resource from the
// Hetzner Cloud backend.
func CheckResourceExists(name string, k BackendKey) resource.TestCheckFunc {
	const op = "testsupport/CheckResourceExists"

	return func(s *terraform.State) error {
		if err := backendResourceByKey(s, name, k); err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
		return nil
	}
}

// CheckResourcesDestroyed checks if resources of resType do not exsist in the
// backend anymore.
func CheckResourcesDestroyed(resType string, k BackendKey) resource.TestCheckFunc {
	const op = "testsupport/CheckResourceDestroyed"

	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != resType {
				continue
			}
			err := backendResourceByKey(s, name, k)
			if err != nil && !errors.Is(err, errMissingInBackend) {
				return fmt.Errorf("%s: %v", op, err)
			}
		}
		return nil
	}
}

var errMissingInBackend = errors.New("missing in backend")

func backendResourceByKey(s *terraform.State, name string, k BackendKey) error {
	const op = "testsupport/backendResourceByKey"

	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return fmt.Errorf("%s: resource %s: not found", op, name)
	}
	if rs.Primary.ID == "" {
		return fmt.Errorf("%s: resource %s: no primary id", op, name)
	}
	id, err := strconv.Atoi(rs.Primary.ID)
	if err != nil {
		return fmt.Errorf("%s: resource %s: primary id: %w", op, name, err)
	}
	client, err := CreateClient()
	if err != nil {
		return fmt.Errorf("%s: create client: %w", op, err)
	}
	if !k(client, id) {
		return fmt.Errorf("%s: resource %s: %w", op, name, errMissingInBackend)
	}
	return nil
}

// CheckResourceAttrFunc uses valueFunc to obtain the expected attribute value.
//
// This allows to delay determining the expected value to just before the
// moment it is checked. In contrast to resource.TestCheckResourceAttrPtr
// valueFunc can return the string representation of any value and is not
// restricted to string pointers.
func CheckResourceAttrFunc(name, key string, valueFunc func() string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		value := valueFunc()
		return resource.TestCheckResourceAttr(name, key, value)(s)
	}
}
