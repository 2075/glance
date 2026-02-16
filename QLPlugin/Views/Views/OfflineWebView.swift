import Cocoa
import os.log
import WebKit

// Block all URLs except those starting with "blob:" or "file://"
let blockRules = """
[
	{
		"trigger": {
			"url-filter": ".*"
		},
		"action": {
			"type": "block"
		}
	},
	{
		"trigger": {
			"url-filter": "blob:.*"
		},
		"action": {
			"type": "ignore-previous-rules"
		}
	},
	{
		"trigger": {
			"url-filter": "file://.*"
		},
		"action": {
			"type": "ignore-previous-rules"
		}
	}
]
"""

/// `WKWebView` which only allows the loading of local resources
class OfflineWebView: WKWebView {
	required init?(coder decoder: NSCoder) {
		super.init(coder: decoder)

		WKContentRuleListStore.default().compileContentRuleList(
			forIdentifier: "ContentBlockingRules",
			encodedContentRuleList: blockRules
		) { contentRuleList, error in
			if let error = error {
				os_log(
					"Error compiling WKWebView content rule list: %{public}s",
					log: Log.render,
					type: .error,
					error.localizedDescription
				)
			} else if let contentRuleList = contentRuleList {
				self.configuration.userContentController.add(contentRuleList)
			} else {
				os_log(
					"Error adding WKWebView content rule list: Content rule list is not defined",
					log: Log.render,
					type: .error
				)
			}
		}
	}

	// MARK: - Keyboard zoom (Cmd+Plus / Cmd+Minus / Cmd+0)

	private static let zoomStep = 1.1
	private static let zoomRange = 0.5...3.0

	override func keyDown(with event: NSEvent) {
		guard event.modifierFlags.contains(.command) else {
			super.keyDown(with: event)
			return
		}
		switch event.charactersIgnoringModifiers {
		case "+", "=":
			pageZoom = min(pageZoom * Self.zoomStep, Self.zoomRange.upperBound)
		case "-":
			pageZoom = max(pageZoom / Self.zoomStep, Self.zoomRange.lowerBound)
		case "0":
			pageZoom = 1.0
		default:
			super.keyDown(with: event)
		}
	}
}
