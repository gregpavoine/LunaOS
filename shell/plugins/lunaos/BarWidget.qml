import QtQuick
import Quickshell.Io
import qs.Ui
import qs.Commons
import qs.Commons as Commons

// LUNA status and goal controls in the bar. Everything goes through
// omarchy-lunaos, which talks to the Core over its owner socket; this file
// holds no credential and grants nothing. Processes start only when the popup
// opens or the owner presses a button, and the follow-up refresh that tracks a
// submitted goal runs only while the popup is open, so an idle shell starts
// nothing (see test/shell.d/shell-idle-spawn-test.sh).
BarWidget {
  id: root
  moduleName: "omarchy.lunaos"

  property bool popupOpen: false
  property bool coreAvailable: false
  property var capabilities: null
  property var goals: []
  property string message: ""
  property string trackedGoal: ""
  property int trackRemaining: 0
  property bool busy: statusProc.running || goalsProc.running || actionProc.running

  readonly property var componentKeys: ["core", "storage", "policy_broker", "local_llm", "hardware"]

  // Shape contract for shell summon/hide/toggle routing.
  readonly property bool opened: popupOpen
  function open() { popupOpen = true }
  function close() { popupOpen = false }
  function toggle() { popupOpen = !popupOpen }

  onPopupOpenChanged: {
    if (popupOpen) refresh()
    else {
      trackedGoal = ""
      trackRemaining = 0
    }
  }

  function refresh() {
    if (!statusProc.running) statusProc.running = true
    if (!goalsProc.running) goalsProc.running = true
  }

  function runAction(args) {
    if (actionProc.running) return
    message = ""
    actionProc.command = ["omarchy-lunaos"].concat(args)
    actionProc.running = true
  }

  function parse(text) {
    try { return JSON.parse(String(text || "")) } catch (e) { return null }
  }

  function failureText(exitCode, stderrText) {
    if (exitCode === 3) return I18n.t("lunaos.coreUnavailable")
    if (exitCode === 4) return I18n.t("lunaos.notAuthorized")
    return String(stderrText || "").trim() || I18n.t("lunaos.failed")
  }

  // Literal keys, so the catalog checks can see every string this file uses.
  function componentLabel(key) {
    switch (key) {
    case "core": return I18n.t("lunaos.component.core")
    case "storage": return I18n.t("lunaos.component.storage")
    case "policy_broker": return I18n.t("lunaos.component.policyBroker")
    case "local_llm": return I18n.t("lunaos.component.localModel")
    case "hardware": return I18n.t("lunaos.component.hardware")
    }
    return key
  }

  function statusLabel(status) {
    switch (status) {
    case "ready": return I18n.t("lunaos.status.ready")
    case "unavailable": return I18n.t("lunaos.status.unavailable")
    case "unverified": return I18n.t("lunaos.status.unverified")
    case "approved": return I18n.t("lunaos.status.approved")
    case "queued": return I18n.t("lunaos.status.queued")
    case "running": return I18n.t("lunaos.status.running")
    case "completed": return I18n.t("lunaos.status.completed")
    case "failed": return I18n.t("lunaos.status.failed")
    case "blocked": return I18n.t("lunaos.status.blocked")
    case "paused": return I18n.t("lunaos.status.paused")
    case "cancelled": return I18n.t("lunaos.status.cancelled")
    }
    return I18n.t("lunaos.status.unknown")
  }

  function componentStatus(key) {
    var entry = capabilities ? capabilities[key] : null
    return entry && entry.status ? entry.status : "unknown"
  }

  function statusColor(status) {
    if (status === "ready" || status === "completed") return Commons.Color.accent
    if (status === "unavailable" || status === "failed" || status === "blocked") return Commons.Color.urgent
    return Commons.Color.muted
  }

  function terminal(status) {
    return status === "completed" || status === "failed" || status === "cancelled" || status === "blocked"
  }

  Process {
    id: statusProc
    command: ["omarchy-lunaos", "status"]
    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: {
        var parsed = root.parse(text)
        if (parsed) root.capabilities = parsed
      }
    }
    onExited: function(exitCode) {
      root.coreAvailable = exitCode === 0
      if (exitCode !== 0) {
        root.capabilities = null
        root.message = root.failureText(exitCode, "")
      }
    }
  }

  Process {
    id: goalsProc
    command: ["omarchy-lunaos", "goals"]
    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: {
        var parsed = root.parse(text)
        root.goals = parsed && parsed.goals ? parsed.goals.slice(0, 8) : []
        if (root.trackedGoal === "") return
        for (var i = 0; i < root.goals.length; i++) {
          if (root.goals[i].id === root.trackedGoal && root.terminal(root.goals[i].status)) root.trackedGoal = ""
        }
      }
    }
  }

  Process {
    id: actionProc
    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: {
        var parsed = root.parse(text)
        if (parsed && parsed.id && !root.terminal(parsed.status)) {
          root.trackedGoal = parsed.id
          root.trackRemaining = 20
        }
      }
    }
    stderr: StdioCollector {
      id: actionErr
      waitForEnd: true
    }
    onExited: function(exitCode) {
      if (exitCode !== 0) root.message = root.failureText(exitCode, actionErr.text)
      root.refresh()
    }
  }

  // Follows a submitted goal to its verified result, only while the owner is
  // looking at the popup and for a bounded number of checks.
  Timer {
    interval: 1500
    repeat: true
    running: root.popupOpen && root.trackedGoal !== "" && root.trackRemaining > 0
    onTriggered: {
      root.trackRemaining -= 1
      if (!goalsProc.running) goalsProc.running = true
    }
  }

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

  BarIconButton {
    id: button
    anchors.fill: parent
    bar: root.bar
    text: ""
    slotSize: Style.bar.statusSlot
    tooltipText: ""
    onPressed: function(b) {
      if (b === Qt.MiddleButton) root.refresh()
      else root.toggle()
    }
  }

  PopupCard {
    id: popup
    anchorItem: root
    bar: root.bar
    owner: root
    open: root.popupOpen
    contentWidth: popup.fittedContentWidth(Style.space(340))
    contentHeight: popup.fittedContentHeight(column.implicitHeight)

    Column {
      id: column
      anchors.fill: parent
      spacing: Style.space(8)

      Text {
        textFormat: Text.PlainText
        text: "LUNA"
        color: Commons.Color.popups.text
        font.family: Style.font.family
        font.pixelSize: Style.font.subtitle
        font.bold: true
      }

      Repeater {
        model: root.coreAvailable ? root.componentKeys : []
        delegate: Row {
          required property string modelData
          width: column.width
          Text {
            textFormat: Text.PlainText
            width: parent.width * 0.55
            text: root.componentLabel(modelData)
            color: Commons.Color.popups.text
            font.family: Style.font.family
            font.pixelSize: Style.font.body
          }
          Text {
            textFormat: Text.PlainText
            width: parent.width * 0.45
            horizontalAlignment: Text.AlignRight
            text: root.statusLabel(root.componentStatus(modelData))
            color: root.statusColor(root.componentStatus(modelData))
            font.family: Style.font.family
            font.pixelSize: Style.font.body
          }
        }
      }

      PanelSeparator { width: column.width; visible: root.coreAvailable }

      Text {
        textFormat: Text.PlainText
        visible: root.coreAvailable
        text: root.goals.length > 0 ? I18n.t("lunaos.goals") : I18n.t("lunaos.noGoals")
        color: Commons.Color.muted
        font.family: Style.font.family
        font.pixelSize: Style.font.bodySmall
      }

      Repeater {
        model: root.coreAvailable ? root.goals : []
        delegate: Row {
          required property var modelData
          width: column.width
          Text {
            textFormat: Text.PlainText
            width: parent.width * 0.65
            elide: Text.ElideRight
            text: modelData.title || modelData.action
            color: Commons.Color.popups.text
            font.family: Style.font.family
            font.pixelSize: Style.font.bodySmall
          }
          Text {
            textFormat: Text.PlainText
            width: parent.width * 0.35
            horizontalAlignment: Text.AlignRight
            text: root.statusLabel(modelData.status)
            color: root.statusColor(modelData.status)
            font.family: Style.font.family
            font.pixelSize: Style.font.bodySmall
          }
        }
      }

      Text {
        textFormat: Text.PlainText
        visible: root.message !== ""
        width: column.width
        wrapMode: Text.Wrap
        text: root.message
        color: Commons.Color.urgent
        font.family: Style.font.family
        font.pixelSize: Style.font.bodySmall
      }

      Row {
        spacing: Style.space(6)
        Button {
          text: I18n.t("lunaos.checksum")
          bordered: true
          enabled: root.coreAvailable && !root.busy
          onClicked: root.runAction(["checksum"])
        }
        Button {
          text: I18n.t("lunaos.stopIntimacy")
          bordered: true
          enabled: root.coreAvailable && !root.busy
          onClicked: root.runAction(["stop-intimacy"])
        }
        Button {
          text: I18n.t("lunaos.refresh")
          bordered: true
          enabled: !root.busy
          onClicked: root.refresh()
        }
      }
    }
  }
}
