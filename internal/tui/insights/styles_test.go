package insights

import "testing"

func TestStylesRenderWithoutPanic(t *testing.T) {
	// Verify heat styles render without panic
	testText := "test"
	for i, style := range HeatStyles {
		result := style.Render(testText)
		if result == "" {
			t.Errorf("HeatStyles[%d] rendered empty string", i)
		}
	}

	// Verify individual heat styles
	_ = Heat0.Render(testText)
	_ = Heat1.Render(testText)
	_ = Heat2.Render(testText)
	_ = Heat3.Render(testText)
	_ = Heat4.Render(testText)

	// Verify stats table styles
	_ = HeaderStyle.Render(testText)
	_ = NameStyle.Render(testText)
	_ = CountStyle.Render(testText)
	_ = PercentStyle.Render(testText)

	// Verify section title style
	_ = SectionTitleStyle.Render(testText)
}

func TestHeatStylesLength(t *testing.T) {
	// Verify we have 5 heat levels (0-4)
	if len(HeatStyles) != 5 {
		t.Errorf("expected 5 heat styles, got %d", len(HeatStyles))
	}
}
