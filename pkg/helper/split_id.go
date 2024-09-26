package helper

import (
	"fmt"
	"strconv"
	strings "strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

func SplitIDToStrings(id string, resource_type string) (string, string, error) {
	parts := strings.Split(id, dbt_cloud.ID_DELIMITER)
	if len(parts) != 2 {

		err := fmt.Errorf(
			"expected ID in the format 'id1%sid2' to import a %s, got: %s",
			dbt_cloud.ID_DELIMITER,
			resource_type,
			id,
		)
		return "", "", err
	}

	return parts[0], parts[1], nil
}

func SplitIDToInts(id string, resource_type string) (int, int, error) {
	parts := strings.Split(id, dbt_cloud.ID_DELIMITER)
	if len(parts) != 2 {

		err := fmt.Errorf(
			"expected ID in the format 'id1%sid2' to import a %s, got: %s",
			dbt_cloud.ID_DELIMITER,
			resource_type,
			id,
		)
		return 0, 0, err
	}

	id1, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("error converting %s to int when splitting the ID", parts[0])
	}

	id2, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("error converting %s to int when splitting the ID", parts[1])
	}

	return id1, id2, nil
}
