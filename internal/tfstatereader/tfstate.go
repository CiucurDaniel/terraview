package tfstatereader

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/CiucurDaniel/terraview/internal/config"
	"github.com/fujiwara/tfstate-lookup/tfstate"
)

// TFStateHandler handles operations on the Terraform state file.
type TFStateHandler struct {
	StateFilePath string
	State         *tfstate.TFState
}

// NewTFStateHandler creates a new TFStateHandler.
func NewTFStateHandler(stateFilePath string) (*TFStateHandler, error) {
	// Create a context
	ctx := context.Background()

	// Read the Terraform state file from the appropriate source
	var state *tfstate.TFState
	var err error
	if isURL(stateFilePath) {
		state, err = tfstate.ReadURL(ctx, stateFilePath)
	} else {
		state, err = tfstate.ReadFile(ctx, stateFilePath)
	}

	if err != nil {
		return nil, fmt.Errorf("error reading tfstate file: %v", err)
	}

	return &TFStateHandler{
		StateFilePath: stateFilePath,
		State:         state,
	}, nil
}

// isURL checks if the provided string is a URL.
func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "s3://") || strings.HasPrefix(path, "gs://") ||
		strings.HasPrefix(path, "azurerm://") || strings.HasPrefix(path, "remote://")
}

// GetImportantAttributes retrieves important attributes for a given resource.
func (h *TFStateHandler) GetImportantAttributes(resource string) ([]string, error) {
	cfg := config.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// List
	//fmt.Println("Listing resources")
	//fmt.Println(h.State.List())

	// Find the resource in the state
	obj, err := h.State.Lookup(resource)
	if err != nil {
		return nil, fmt.Errorf("resource %s not found in tfstate: %v", resource, err)
	}

	attributesMap := obj.Value.(map[string]interface{})

	// Extract important attributes based on config
	resourceType := strings.Split(resource, ".")[0]

	var importantAttrs []string
	for _, resConfig := range cfg.ImportantAttributes {
		if resConfig.Name == resourceType {
			for _, attr := range resConfig.Attributes {
				// Directly access the attribute value from the attributes map
				if value, ok := attributesMap[attr]; ok {
					importantAttrs = append(importantAttrs, fmt.Sprintf("%s: %s", attr, value))
				} else {
					log.Printf("Attribute %s not found for resource %s", attr, resource)
				}
			}
			break
		}
	}

	if len(importantAttrs) == 0 {
		return nil, fmt.Errorf("no important attributes found for resource %s", resource)
	}

	return importantAttrs, nil
}

// IsCreatedWithList checks if a resource was created with a count or for_each.
func (h *TFStateHandler) IsCreatedWithList(resource string) bool {
	resourceList, err := h.State.List()
	if err != nil {
		log.Printf("error listing resources: %v", err)
		return false
	}

	for _, res := range resourceList {
		if strings.HasPrefix(res, resource+"[") || strings.HasPrefix(res, resource+"[\"") {
			return true
		}
	}
	return false
}

// GetListOfNamesForResource returns the list of actual names for a resource created with count or for_each.
func (h *TFStateHandler) GetListOfNamesForResource(resource string) ([]string, error) {
	resourceList, err := h.State.List()
	if err != nil {
		return nil, fmt.Errorf("error listing resources: %v", err)
	}

	var resourceNames []string
	for _, res := range resourceList {
		if strings.HasPrefix(res, resource+"[") || strings.HasPrefix(res, resource+"[\"") {
			resourceNames = append(resourceNames, res)
		}
	}

	if len(resourceNames) == 0 {
		return nil, fmt.Errorf("no resources found for %s", resource)
	}

	return resourceNames, nil
}
