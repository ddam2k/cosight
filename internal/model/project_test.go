package model

import "testing"

func TestCreateProjectNormalizeAndValidate(t *testing.T) {
	description := "  source explorer  "
	path := "  /repositories/cosight  "
	input := CreateProject{
		Scope:                ProjectScopePersonal,
		Name:                 " Cosight ",
		Slug:                 "cosight",
		Description:          &description,
		RepositoryAlias:      " main ",
		RepositorySourceType: RepositoryLocalPath,
		RepositoryPath:       &path,
		ExcludePatterns:      []string{" dist/** ", ""},
	}
	input.Normalize()
	if fieldErrors := input.Validate(); len(fieldErrors) != 0 {
		t.Fatalf("unexpected field errors: %#v", fieldErrors)
	}
	if input.Name != "Cosight" || input.RepositoryPath == nil || *input.RepositoryPath != "/repositories/cosight" {
		t.Fatalf("input was not normalized: %#v", input)
	}
	if len(input.ExcludePatterns) != 1 || input.ExcludePatterns[0] != "dist/**" {
		t.Fatalf("patterns were not normalized: %#v", input.ExcludePatterns)
	}
}

func TestCreateProjectValidatesScopeAndRepository(t *testing.T) {
	organizationID := "organization-id"
	path := "relative/path"
	input := CreateProject{
		Scope:                ProjectScopePersonal,
		OrganizationID:       &organizationID,
		Name:                 "Cosight",
		Slug:                 "INVALID",
		RepositoryAlias:      "main",
		RepositorySourceType: RepositoryLocalPath,
		RepositoryPath:       &path,
	}
	fieldErrors := input.Validate()
	for _, field := range []string{"organizationId", "slug", "repositoryPath"} {
		if fieldErrors[field] == "" {
			t.Fatalf("expected %s validation error: %#v", field, fieldErrors)
		}
	}
}
