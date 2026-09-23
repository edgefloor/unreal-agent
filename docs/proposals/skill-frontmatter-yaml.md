# Decode skill frontmatter as YAML

Status: proposal only. No implementation or tests are included.

## Problem

`parseSkillFrontmatter` in `harness/tool/registry.go` splits lines at the first
colon and trims whitespace. It does not decode YAML scalar syntax.

At base commit `df8b0ba`, discovery retained quote characters in `name: "review"`,
returned the literal marker `>-` for a folded description, and included an inline
comment in the registered name. These values change skill lookup names or remove
the description used to select a skill.

## Proposed scope

Decode the frontmatter block into the existing name and description fields using
a YAML parser. Include the selected dependency in the eventual implementation
commit. Preserve unrelated metadata, existing field validation, duplicate-name
checks, and discovery continuing after an invalid file. Do not change skill
execution or the session storage format.

## Acceptance criteria

- Quoted names and descriptions decode to their string values.
- Folded and literal descriptions retain their YAML-defined text.
- Inline comments do not become part of the registered name.
- Malformed metadata produces a useful discovery error.
- Unrelated metadata remains accepted, and duplicate names remain rejected.
- A discovered skill can be resolved by its decoded name through `SkillUse`.

## Planned validation

Add table-driven discovery tests for quoted, folded, literal, commented, and
malformed frontmatter. Verify discovery followed by `SkillUse` lookup and preserve
existing invalid-file and duplicate-name coverage. Run `make test check build`
when implementation is authorized.
