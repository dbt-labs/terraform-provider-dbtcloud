#!/usr/bin/env python3
"""Add ResourceMetadata to resource models and resource_metadata to resource schemas."""
import os
import re

OBJECTS_DIR = "pkg/framework/objects"
RESOURCE_PACKAGES = [
    "connection_catalog_config", "environment_variable", "environment_variable_job_override",
    "extended_attributes", "fabric_credential", "ip_restrictions_rule", "license_map",
    "lineage_integration", "model_notifications", "notification", "oauth_configuration",
    "platform_metadata_credentials", "postgres_credential", "project", "project_artefacts",
    "redshift_credential", "salesforce_credential", "scim_group_partial_permissions",
    "scim_group_permissions", "snowflake_credential", "spark_credential", "starburst_credential",
    "synapse_credential", "group_partial_permissions",
]

def add_to_model(pkg: str) -> bool:
    path = os.path.join(OBJECTS_DIR, pkg, "model.go")
    if not os.path.isfile(path):
        return False
    with open(path) as f:
        content = f.read()
    if 'ResourceMetadata types.Dynamic' in content or 'resource_metadata' in content:
        return False
    # Find last field of a struct that looks like *ResourceModel
    # Insert before the closing } of the Resource model struct (first struct with "ResourceModel" in name that has tfsdk)
    lines = content.split("\n")
    new_lines = []
    i = 0
    in_resource_model = False
    last_brace_pos = -1
    while i < len(lines):
        line = lines[i]
        if "ResourceModel struct" in line or "ResourceModel struct {" in line:
            in_resource_model = True
            last_brace_pos = -1
        if in_resource_model and line.strip() == "}":
            # Found closing brace of resource model - insert before it
            indent = len(line) - len(line.lstrip())
            new_lines.append(" " * indent + 'ResourceMetadata types.Dynamic `tfsdk:"resource_metadata"`')
            in_resource_model = False
        new_lines.append(line)
        i += 1
    new_content = "\n".join(new_lines)
    if new_content == content:
        return False
    with open(path, "w") as f:
        f.write(new_content)
    return True

def add_to_schema(pkg: str) -> bool:
    path = os.path.join(OBJECTS_DIR, pkg, "schema.go")
    if not os.path.isfile(path):
        return False
    with open(path) as f:
        content = f.read()
    if '"resource_metadata"' in content:
        return False
    # Add resource_metadata attribute before the closing of resource Attributes map
    # Pattern: "some_attr": resource_schema.XAttribute{ ... }, then }, then }
    # We need to add before "		},\n	}" or "	},\n\t}\n}" for resource schema
    if "resource_schema." in content:
        # Find last resource_schema attribute and add after it
        insert = '''			"resource_metadata": resource_schema.DynamicAttribute{
				Optional:    true,
				Description: "Optional migration identity metadata persisted in Terraform state.",
			},
		},
	}'''
        if "resource_schema.DynamicAttribute" in content and "resource_metadata" in content:
            return False
        # Insert before first occurrence of "\n\t},\n}" or "\n\t\t},\n\t}" that closes resource Attributes
        pattern = re.compile(r'(\s+)"(\w+)": resource_schema\.\w+Attribute\{\s*[^}]+\},\s*\n(\s+)\},\s*\n(\s+)\}')
        match = pattern.search(content)
        if not match:
            return False
        # Add before the closing "},\n\t}"
        old = content
        idx = content.rfind('},\n\t}')
        if idx == -1:
            idx = content.rfind("},\n\t}\n}")
        if idx == -1:
            return False
        before = content[:idx]
        after = content[idx:]
        addition = '''			"resource_metadata": resource_schema.DynamicAttribute{
				Optional:    true,
				Description: "Optional migration identity metadata persisted in Terraform state.",
			},
'''
        new_content = before + addition + after
        if new_content == content:
            return False
        with open(path, "w") as f:
            f.write(new_content)
        return True
    elif "schema." in content and "schema.Attribute" in content:
        # schema.Schema style
        insert = '''			"resource_metadata": schema.DynamicAttribute{
				Optional:    true,
				Description: "Optional migration identity metadata persisted in Terraform state.",
			},
		},
	}'''
        idx = content.find("Attributes: map[string]schema.Attribute{")
        if idx == -1:
            return False
        close_idx = content.rfind("},\n\t}\n}", idx)
        if close_idx == -1:
            close_idx = content.rfind("},\n\t}", idx)
        if close_idx == -1:
            return False
        before = content[:close_idx]
        after = content[close_idx:]
        addition = '''			"resource_metadata": schema.DynamicAttribute{
				Optional:    true,
				Description: "Optional migration identity metadata persisted in Terraform state.",
			},
'''
        new_content = before + addition + after
        if new_content == content:
            return False
        with open(path, "w") as f:
            f.write(new_content)
        return True
    return False

if __name__ == "__main__":
    os.chdir(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    for pkg in RESOURCE_PACKAGES:
        if add_to_model(pkg):
            print(f"model: {pkg}")
        if add_to_schema(pkg):
            print(f"schema: {pkg}")
