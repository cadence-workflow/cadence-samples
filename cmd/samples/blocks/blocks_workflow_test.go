package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeResponse(t *testing.T) {
	// Test with empty votes
	t.Run("EmptyVotes", func(t *testing.T) {
		votes := []map[string]string{}
		result := makeResponse(votes)

		// Verify the structure
		assert.Equal(t, "formattedData", result.CadenceResponseType)
		assert.Equal(t, "blocks", result.Format)
		assert.Len(t, result.Blocks, 5)

		// Verify JSON serialization matches expected output
		resultJSON, err := json.Marshal(result)
		assert.NoError(t, err)

		expectedJSON := `{
  "cadenceResponseType": "formattedData",
  "format": "blocks",
  "blocks": [
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "## Lunch options\nWe're voting on where to order lunch today. Select the option you want to vote for."
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "| lunch order vote | meal | requests |\n|-------|-------|-------|\n| No votes yet |\n"
      }
    },
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "|  Picture |  Description  |\n|---|----|\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Red_roast_duck_curry.jpg/200px-Red_roast_duck_curry.jpg) | Farmhouse - Red Thai Curry: (Thai: แกง, romanized: kaeng, pronounced [kɛ̄ːŋ]) is a dish in Thai cuisine made from curry paste, coconut milk or water, meat, seafood, vegetables or fruit, and herbs. Curries in Thailand mainly differ from the Indian subcontinent in their use of ingredients such as fresh rhizomes, herbs, and aromatic leaves rather than a mix of dried spices. |\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/0/0c/B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png/200px-B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png) | Ler Ros: Lemongrass Tofu Bahn Mi: In Vietnamese cuisine, bánh mì, bánh mỳ or banh mi is a sandwich consisting of a baguette filled with various ingredients, most commonly including a protein such as pâté, chicken, or pork, and vegetables such as lettuce, cilantro, and cucumber. |\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/5/54/Ethiopian_wat.jpg/960px-Ethiopian_wat.jpg) | Ethiopian Wat: Wat is a traditional Ethiopian dish made from a mixture of spices, vegetables, and legumes. It is typically served with injera, a sourdough flatbread that is used to scoop up the food. |\n\n\n\n(source wikipedia)"
      }
    },
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Farmhouse"
          },
          "action": {
            "type": "signal",
            "signal_name": "lunch_order",
            "signal_value": {
              "location": "farmhouse - red thai curry",
              "requests": "spicy"
            }
          }
        },
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Ethiopian"
          },
          "action": {
            "type": "signal",
            "signal_name": "no_lunch_order_walk_in_person",
            "workflow_id": "in-person-order-workflow"
          }
        },
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Ler Ros"
          },
          "action": {
            "type": "signal",
            "signal_name": "lunch_order",
            "signal_value": {
              "location": "Ler Ros",
              "meal": "tofo Bahn Mi"
            }
          }
        }
      ]
    }
  ]
}`

		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	// Test with some votes
	t.Run("WithVotes", func(t *testing.T) {
		votes := []map[string]string{
			{"location": "farmhouse - red thai curry", "requests": "spicy"},
			{"location": "Ler Ros", "meal": "tofo Bahn Mi"},
		}
		result := makeResponse(votes)

		// Verify the structure
		assert.Equal(t, "formattedData", result.CadenceResponseType)
		assert.Equal(t, "blocks", result.Format)
		assert.Len(t, result.Blocks, 5)

		// Verify JSON serialization matches expected output
		resultJSON, err := json.Marshal(result)
		assert.NoError(t, err)

		expectedJSON := `{
  "cadenceResponseType": "formattedData",
  "format": "blocks",
  "blocks": [
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "## Lunch options\nWe're voting on where to order lunch today. Select the option you want to vote for."
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "| lunch order vote | meal | requests |\n|-------|-------|-------|\n| farmhouse - red thai curry |  | spicy |\n| Ler Ros | tofo Bahn Mi |  |\n"
      }
    },
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "|  Picture |  Description  |\n|---|----|\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Red_roast_duck_curry.jpg/200px-Red_roast_duck_curry.jpg) | Farmhouse - Red Thai Curry: (Thai: แกง, romanized: kaeng, pronounced [kɛ̄ːŋ]) is a dish in Thai cuisine made from curry paste, coconut milk or water, meat, seafood, vegetables or fruit, and herbs. Curries in Thailand mainly differ from the Indian subcontinent in their use of ingredients such as fresh rhizomes, herbs, and aromatic leaves rather than a mix of dried spices. |\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/0/0c/B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png/200px-B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png) | Ler Ros: Lemongrass Tofu Bahn Mi: In Vietnamese cuisine, bánh mì, bánh mỳ or banh mi is a sandwich consisting of a baguette filled with various ingredients, most commonly including a protein such as pâté, chicken, or pork, and vegetables such as lettuce, cilantro, and cucumber. |\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/5/54/Ethiopian_wat.jpg/960px-Ethiopian_wat.jpg) | Ethiopian Wat: Wat is a traditional Ethiopian dish made from a mixture of spices, vegetables, and legumes. It is typically served with injera, a sourdough flatbread that is used to scoop up the food. |\n\n\n\n(source wikipedia)"
      }
    },
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Farmhouse"
          },
          "action": {
            "type": "signal",
            "signal_name": "lunch_order",
            "signal_value": {
              "location": "farmhouse - red thai curry",
              "requests": "spicy"
            }
          }
        },
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Ethiopian"
          },
          "action": {
            "type": "signal",
            "signal_name": "no_lunch_order_walk_in_person",
            "workflow_id": "in-person-order-workflow"
          }
        },
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Ler Ros"
          },
          "action": {
            "type": "signal",
            "signal_name": "lunch_order",
            "signal_value": {
              "location": "Ler Ros",
              "meal": "tofo Bahn Mi"
            }
          }
        }
      ]
    }
  ]
}`

		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})
}
