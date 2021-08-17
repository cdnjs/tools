package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/pkg/errors"
)

// Unmarshals the human-readable JSON into a *Package,
// setting the legacy `author` field if needed.
func ReadHumanJSONBytes(ctx context.Context, file string, bytes []byte, validateSchema bool) (*Package, error) {
	if validateSchema {
		// validate against human readable JSON schema
		res, err := HumanReadableSchema.Validate(gojsonschema.NewBytesLoader(bytes))
		if err != nil {
			// invalid JSON
			return nil, errors.Wrapf(err, "failed to parse %s", file)
		}

		if !res.Valid() {
			// invalid schema, so return result and custom error
			return nil, InvalidSchemaError{res}
		}
	}

	// unmarshal JSON into package
	var p Package
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", file)
	}

	// if `authors` exists, parse `author` field
	if p.Authors != nil {
		author := parseAuthor(p.Authors)
		p.Author = &author
	}

	p.ctx = ctx
	return &p, nil
}

// ReadNonHumanJSONBytes unmarshals bytes into a *Package,
// validating against the non-human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
func ReadNonHumanJSONBytes(ctx context.Context, name string, bytes []byte) (*Package, error) {
	// validate the non-human readable JSON schema
	res, err := NonHumanReadableSchema.Validate(gojsonschema.NewBytesLoader(bytes))
	if err != nil {
		// invalid JSON
		return nil, errors.Wrapf(err, "failed to parse %s", name)
	}

	if !res.Valid() {
		// invalid schema, so return result and custom error
		return nil, InvalidSchemaError{res}
	}

	var p Package
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", name)
	}

	// schema is valid, but we still need to ensure there are either
	// both `author` and `authors` fields or neither
	authorsNil, authorNil := p.Authors == nil, p.Author == nil
	if authorsNil != authorNil {
		return nil, errors.Wrapf(err, "`author` and `authors` must be either both nil or both non-nil - %s", name)
	}

	if !authorsNil {
		// `authors` exists, so need to verify `author` is parsed correctly
		author := *p.Author
		parsedAuthor := parseAuthor(p.Authors)
		if author != parsedAuthor {
			return nil, fmt.Errorf("author parse: actual `%s` != expected `%s`", author, parsedAuthor)
		}
	}

	p.ctx = ctx
	return &p, nil
}

// If `authors` exists, we need to parse `author` field
// for legacy compatibility with API.
func parseAuthor(authors []Author) string {
	var authorStrings []string
	for _, author := range authors {
		authorString := *author.Name
		if author.Email != nil {
			authorString += fmt.Sprintf(" <%s>", *author.Email)
		}
		if author.URL != nil {
			authorString += fmt.Sprintf(" (%s)", *author.URL)
		}
		authorStrings = append(authorStrings, authorString)
	}
	return strings.Join(authorStrings, ",")
}
