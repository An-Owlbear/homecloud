package templates

import (
	"maps"

	kratos "github.com/ory/kratos-client-go"
)

func boolPointer(b bool) *bool {
	return &b
}

func getUiContainerMethod(ui kratos.UiContainer) string {
	if ui.Method != "" {
		return ui.Method
	}
	return "POST"
}

func getInputNodeValue(node kratos.UiNode) string {
	if node.Attributes.UiNodeInputAttributes == nil {
		return ""
	}

	if value, ok := node.Attributes.UiNodeInputAttributes.Value.(string); ok {
		return value
	}

	return ""
}

func getInputNodeAutocomplete(node kratos.UiNode) string {
	if node.Attributes.UiNodeInputAttributes.Autocomplete == nil {
		return "off"
	}
	return *node.Attributes.UiNodeInputAttributes.Autocomplete
}

// organiseForms organises the nodes into their separate forms
func organiseForms(flow kratos.UiContainer) (forms map[string][]kratos.UiNode) {
	// Sorts nodes into forms and organises default nodes
	forms = map[string][]kratos.UiNode{}
	var defaultNodes []kratos.UiNode
	for _, node := range flow.Nodes {
		if node.Group != "default" {
			forms[node.Group] = append(forms[node.Group], node)
		} else {
			defaultNodes = append(defaultNodes, node)
		}
	}

	// Adds default nodes to all forms
	for _, node := range defaultNodes {
		for key := range maps.Keys(forms) {
			forms[key] = append(forms[key], node)
		}
	}

	return
}
