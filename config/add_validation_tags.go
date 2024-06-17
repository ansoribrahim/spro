package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

type OpenAPISpec struct {
	Components struct {
		Schemas map[string]struct {
			Required   []string `yaml:"required"`
			Properties map[string]struct {
				Type        string `yaml:"type"`
				Minimum     int    `yaml:"minimum,omitempty"`
				Maximum     int    `yaml:"maximum,omitempty"`
				Description string `yaml:"description,omitempty"`
			} `yaml:"properties"`
		} `yaml:"schemas"`
	} `yaml:"components"`
}

func loadOpenAPISpec(filePath string) (*OpenAPISpec, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	err = yaml.Unmarshal(data, &spec)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

func addValidationTags(spec *OpenAPISpec, inputFilePath, outputFilePath string) error {
	input, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		return err
	}

	content := string(input)

	for _, schema := range spec.Components.Schemas {
		for propertyName, property := range schema.Properties {
			var validationTags []string

			// Check for required fields
			for _, requiredField := range schema.Required {
				if requiredField == propertyName {
					validationTags = append(validationTags, "required")
					break
				}
			}

			// Check for minimum and maximum
			if property.Minimum != 0 {
				validationTags = append(validationTags, fmt.Sprintf("gte=%d", property.Minimum))
			}
			if property.Maximum != 0 {
				validationTags = append(validationTags, fmt.Sprintf("lte=%d", property.Maximum))
			}

			if len(validationTags) > 0 {
				validationTag := strings.Join(validationTags, ",")
				jsonTag := fmt.Sprintf("json:\"%s\"", propertyName)
				newTag := fmt.Sprintf("%s validate:\"%s\"", jsonTag, validationTag)
				oldTag := fmt.Sprintf("`%s`", jsonTag)
				content = strings.ReplaceAll(content, oldTag, fmt.Sprintf("`%s`", newTag))
			}
		}
	}

	return ioutil.WriteFile(outputFilePath, []byte(content), 0644)
}

func main() {
	// File paths
	openAPISpecPath := "api.yml"
	inputFilePath := "generated/api.gen.go"
	outputFilePath := "generated/api.gen.go"

	// Load OpenAPI spec
	spec, err := loadOpenAPISpec(openAPISpecPath)
	if err != nil {
		log.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	// Add validation tags to generated code
	err = addValidationTags(spec, inputFilePath, outputFilePath)
	if err != nil {
		log.Fatalf("Failed to add validation tags: %v", err)
	}

	log.Println("Successfully added validation tags to the generated code.")
}
