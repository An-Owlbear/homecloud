package templates

import (
	kratos "github.com/ory/kratos-client-go"
)

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
