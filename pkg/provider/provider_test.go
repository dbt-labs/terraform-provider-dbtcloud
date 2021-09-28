package provider_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/provider"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	r := require.New(t)
	err := provider.Provider().InternalValidate()
	r.NoError(err)
}

func TestAuthentication(t *testing.T) {

}
