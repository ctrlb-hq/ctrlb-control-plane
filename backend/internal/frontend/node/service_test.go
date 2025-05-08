package frontendnode

import (
	"errors"
	"reflect"
	"testing"
)

// MockFrontendNodeRepository is a manual mock for FrontendNodeRepositoryInterface
type MockFrontendNodeRepository struct {
	GetComponentsFunc            func(componentType string) (*[]ComponentInfo, error)
	GetComponentSchemaByNameFunc func(componentName string) (any, error)
}

func (m *MockFrontendNodeRepository) GetComponents(componentType string) (*[]ComponentInfo, error) {
	return m.GetComponentsFunc(componentType)
}

func (m *MockFrontendNodeRepository) GetComponentSchemaByName(componentName string) (any, error) {
	return m.GetComponentSchemaByNameFunc(componentName)
}

// TestFrontendNodeService_GetComponents tests the GetComponents method
func TestFrontendNodeService_GetComponents(t *testing.T) {
	mockRepo := &MockFrontendNodeRepository{
		GetComponentsFunc: func(componentType string) (*[]ComponentInfo, error) {
			return &[]ComponentInfo{
				{Name: "ComponentA", Type: "source"},
				{Name: "ComponentB", Type: "processor"},
			}, nil
		},
	}

	service := NewFrontendNodeService(mockRepo)

	result, err := service.GetComponents("source")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expected := &[]ComponentInfo{
		{Name: "ComponentA", Type: "source"},
		{Name: "ComponentB", Type: "processor"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

// TestFrontendNodeService_GetComponentSchemaByName tests the GetComponentSchemaByName method
func TestFrontendNodeService_GetComponentSchemaByName(t *testing.T) {
	mockRepo := &MockFrontendNodeRepository{
		GetComponentSchemaByNameFunc: func(componentName string) (any, error) {
			if componentName == "ComponentA" {
				return map[string]any{"field1": "string", "field2": "int"}, nil
			}
			return nil, errors.New("component not found")
		},
	}

	service := NewFrontendNodeService(mockRepo)

	// Positive case
	schema, err := service.GetComponentSchemaByName("ComponentA")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedSchema := map[string]any{"field1": "string", "field2": "int"}
	if !reflect.DeepEqual(schema, expectedSchema) {
		t.Errorf("expected %+v, got %+v", expectedSchema, schema)
	}

	// Negative case
	_, err = service.GetComponentSchemaByName("UnknownComponent")
	if err == nil {
		t.Fatal("expected an error for unknown component, got nil")
	}
}
