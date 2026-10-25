pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import Quickshell
import Quickshell.Io
import Quickshell.Wayland

ShellRoot {
  id: root

  property bool dashboardVisible: false
  property bool apiOnline: false
  property bool requestActive: false
  property string activePath: ""
  property string activeMethod: ""
  property string pendingRequest: ""
  property string apiMessage: "Connecting to Luna Core"
  property string privacyMessage: "Privacy state unavailable"
  property var goals: []
  property var responseObject: null
  property int responseStatus: 0
  property string clockText: Qt.formatDateTime(new Date(), "HH:mm")

  function request(method, path, body, idempotencyKey) {
    if (requestActive) return
    activePath = path
    activeMethod = method
    responseObject = null
    responseStatus = 0
    var serialized = body === null ? "" : JSON.stringify(body)
    var raw = method + " " + path + " HTTP/1.1\r\nHost: lunaos\r\nConnection: close\r\n"
    if (body !== null) {
      raw += "Content-Type: application/json\r\nContent-Length: " + serialized.length + "\r\n"
    }
    if (idempotencyKey) raw += "Idempotency-Key: " + idempotencyKey + "\r\n"
    raw += "\r\n" + serialized
    pendingRequest = raw
    requestActive = true
    requestTimeout.restart()
    apiSocket.connected = true
  }

  function refresh() {
    if (requestActive) return
    request("GET", "/v1/goals", null, "")
  }

  function runLocalEvidenceCheck() {
    request("POST", "/v1/goals", ({ action: "local.checksum", privacy_class: "general" }), "shell-" + Date.now())
  }

  function stopIntimacy() {
    request("POST", "/v1/intimacy/stop", ({}), "")
  }

  function handleLine(line) {
    var text = String(line).trim()
    if (text.length === 0) return
    if (text.indexOf("HTTP/") === 0) {
      const parts = text.split(" ")
      responseStatus = parts.length > 1 ? Number(parts[1]) : 0
      return
    }
    if (text.charAt(0) === "{" || text.charAt(0) === "[") {
      try {
        responseObject = JSON.parse(text)
      } catch (error) {
        failRequest("Luna Core returned invalid JSON")
        return
      }
      finishRequest()
    }
  }

  function finishRequest() {
    if (!requestActive) return
    requestTimeout.stop()
    apiOnline = responseStatus >= 200 && responseStatus < 300
    if (!apiOnline) {
      const code = responseObject && responseObject.code ? responseObject.code : "HTTP " + responseStatus
      apiMessage = "Core request blocked: " + code
    } else if (activePath === "/v1/goals" && activeMethod === "GET") {
      goals = responseObject && responseObject.goals ? responseObject.goals : []
      apiMessage = "Goal Engine connected"
    } else if (activePath === "/v1/goals") {
      apiMessage = responseObject && responseObject.evidence && responseObject.evidence.verified
        ? "Checksum completed with verified evidence"
        : "Checksum queued; waiting for scheduler"
      refreshTimer.restart()
    } else if (activePath === "/v1/intimacy/stop") {
      privacyMessage = "Intimate processing stopped"
      apiMessage = "Stop state saved"
    } else if (activePath === "/v1/intimacy/relationship/state") {
      privacyMessage = responseObject && responseObject.stopped
        ? "Intimate processing stopped"
        : "No sensitive processing is enabled"
      apiMessage = "Privacy state connected"
    }
    requestActive = false
    apiSocket.connected = false
    requestPump.restart()
  }

  function failRequest(message) {
    if (!requestActive) return
    requestTimeout.stop()
    requestActive = false
    apiOnline = false
    apiMessage = message
    apiSocket.connected = false
    requestPump.restart()
  }

  function nextRequest() {
    if (requestActive || apiSocket.connected) return
    if (activePath === "/v1/goals" && activeMethod === "GET") {
      request("GET", "/v1/intimacy/relationship/state", null, "")
    } else if (activePath === "/v1/intimacy/relationship/state") {
      refreshTimer.restart()
    }
  }

  IpcHandler {
    target: "luna"
    function toggleDashboard() {
      root.dashboardVisible = !root.dashboardVisible
      if (root.dashboardVisible) root.refresh()
    }
  }

  Socket {
    id: apiSocket
    path: "/run/lunaos/core.sock"
    connected: false
    parser: SplitParser {
      splitMarker: "\n"
      onRead: data => root.handleLine(data)
    }
    onConnectedChanged: {
      if (connected && root.pendingRequest.length > 0) {
        write(root.pendingRequest)
        flush()
        root.pendingRequest = ""
      }
    }
    onError: error => root.failRequest("Core socket unavailable")
  }

  Timer {
    id: requestTimeout
    interval: 4000
    onTriggered: root.failRequest("Core API timed out")
  }

  Timer {
    id: requestPump
    interval: 20
    onTriggered: root.nextRequest()
  }

  Timer {
    id: refreshTimer
    interval: 12000
    onTriggered: root.refresh()
  }

  Timer {
    interval: 30000
    running: true
    repeat: true
    onTriggered: root.clockText = Qt.formatDateTime(new Date(), "HH:mm")
  }

  Component.onCompleted: refresh()

  PanelWindow {
    anchors { top: true; left: true; right: true }
    implicitHeight: 56
    color: "transparent"
    exclusionMode: ExclusionMode.Normal
    WlrLayershell.layer: WlrLayer.Top
    WlrLayershell.namespace: "lunaos-panel"

    Rectangle {
      anchors.fill: parent
      color: "#111319"
      border.color: "#2b3039"
      border.width: 1

      RowLayout {
        anchors.fill: parent
        anchors.leftMargin: 22
        anchors.rightMargin: 22
        spacing: 18

        Rectangle {
          Layout.preferredWidth: 30
          Layout.preferredHeight: 30
          radius: 10
          color: "#d1a86b"
          Text { anchors.centerIn: parent; text: "L"; color: "#171717"; font.pixelSize: 17; font.weight: Font.DemiBold }
        }
        Text { text: "LUNAOS"; color: "#f1eee7"; font.pixelSize: 14; font.weight: Font.DemiBold; font.letterSpacing: 2 }
        Rectangle { Layout.fillWidth: true; color: "transparent" }
        Rectangle {
          Layout.preferredWidth: 9
          Layout.preferredHeight: 9
          radius: 5
          color: root.apiOnline ? "#78c6a3" : "#cc8c67"
        }
        Text { text: root.apiOnline ? "CORE CONNECTED" : "CORE OFFLINE"; color: "#aeb4bf"; font.pixelSize: 10; font.letterSpacing: 1 }
        Text { text: root.clockText; color: "#f1eee7"; font.pixelSize: 13; font.family: "monospace" }
        Button {
          id: commandCenterButton
          text: root.dashboardVisible ? "Close" : "Command center"
          onClicked: { root.dashboardVisible = !root.dashboardVisible; if (root.dashboardVisible) root.refresh() }
          contentItem: Text { text: commandCenterButton.text; color: "#171717"; font.pixelSize: 12; font.weight: Font.DemiBold; horizontalAlignment: Text.AlignHCenter; verticalAlignment: Text.AlignVCenter }
          background: Rectangle { radius: 9; color: commandCenterButton.down ? "#bc925a" : "#d1a86b" }
          Layout.preferredHeight: 34
          Layout.preferredWidth: 142
        }
      }
    }
  }

  FloatingWindow {
    title: "LunaOS Command Center"
    visible: root.dashboardVisible
    width: 920
    height: 690
    color: "#111319"
    onVisibleChanged: if (visible) root.refresh()

    Rectangle {
      anchors.fill: parent
      color: "#111319"

      ColumnLayout {
        anchors.fill: parent
        anchors.margins: 30
        spacing: 20

        RowLayout {
          Layout.fillWidth: true
          ColumnLayout {
            spacing: 4
            Text { text: "Command center"; color: "#f1eee7"; font.pixelSize: 28; font.weight: Font.DemiBold }
            Text { text: "Local services and verified work"; color: "#969daa"; font.pixelSize: 13 }
          }
          Item { Layout.fillWidth: true }
          Text { text: root.apiMessage; color: root.apiOnline ? "#78c6a3" : "#cc8c67"; font.pixelSize: 12 }
        }

        RowLayout {
          Layout.fillWidth: true
          spacing: 14
          Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 100
            radius: 16
            color: "#1a1e26"
            border.color: "#2b3039"
            Column {
              anchors.left: parent.left; anchors.verticalCenter: parent.verticalCenter; anchors.leftMargin: 18; spacing: 7
              Text { text: "GOAL ENGINE"; color: "#989eaa"; font.pixelSize: 10; font.letterSpacing: 1.4 }
              Text { text: root.apiOnline ? "Ready" : "Unavailable"; color: root.apiOnline ? "#78c6a3" : "#cc8c67"; font.pixelSize: 20; font.weight: Font.Medium }
            }
          }
          Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 100
            radius: 16
            color: "#1a1e26"
            border.color: "#2b3039"
            Column {
              anchors.left: parent.left; anchors.verticalCenter: parent.verticalCenter; anchors.leftMargin: 18; spacing: 7
              Text { text: "PRIVACY CONTROL"; color: "#989eaa"; font.pixelSize: 10; font.letterSpacing: 1.4 }
              Text { text: root.privacyMessage; color: "#f1eee7"; font.pixelSize: 14; font.weight: Font.Medium }
            }
          }
        }

        RowLayout {
          Layout.fillWidth: true
          Text { text: "Recent goals"; color: "#f1eee7"; font.pixelSize: 17; font.weight: Font.DemiBold }
          Item { Layout.fillWidth: true }
          Button {
            id: checksumButton
            text: "Run local evidence check"
            enabled: root.apiOnline && !root.requestActive
            onClicked: root.runLocalEvidenceCheck()
            contentItem: Text { text: checksumButton.text; color: "#171717"; font.pixelSize: 12; font.weight: Font.DemiBold; horizontalAlignment: Text.AlignHCenter; verticalAlignment: Text.AlignVCenter }
            background: Rectangle { radius: 9; color: checksumButton.enabled ? (checksumButton.down ? "#bc925a" : "#d1a86b") : "#55575b" }
            Layout.preferredHeight: 38
            Layout.preferredWidth: 190
          }
        }

        Rectangle {
          Layout.fillWidth: true
          Layout.fillHeight: true
          radius: 16
          color: "#171a21"
          border.color: "#292e37"
          ColumnLayout {
            anchors.fill: parent
            anchors.margins: 16
            spacing: 8
            RowLayout {
              Layout.fillWidth: true
              Text { text: "ACTION"; color: "#858c99"; font.pixelSize: 9; font.letterSpacing: 1.2; Layout.fillWidth: true }
              Text { text: "STATE"; color: "#858c99"; font.pixelSize: 9; font.letterSpacing: 1.2; Layout.preferredWidth: 115 }
              Text { text: "EVIDENCE"; color: "#858c99"; font.pixelSize: 9; font.letterSpacing: 1.2; Layout.preferredWidth: 100 }
            }
            Rectangle { Layout.fillWidth: true; Layout.preferredHeight: 1; color: "#2a2e37" }
            ListView {
              Layout.fillWidth: true
              Layout.fillHeight: true
              clip: true
              model: root.goals
              spacing: 3
              delegate: Rectangle {
                id: goalRow
                required property int index
                required property var modelData
                width: ListView.view.width
                height: 52
                radius: 9
                color: goalRow.index % 2 === 0 ? "#1b1f27" : "transparent"
                RowLayout {
                  anchors.fill: parent
                  anchors.leftMargin: 12
                  anchors.rightMargin: 12
                  Text { text: goalRow.modelData.action || "local action"; color: "#e8e5de"; font.pixelSize: 12; Layout.fillWidth: true }
                  Text { text: goalRow.modelData.status || "unknown"; color: goalRow.modelData.status === "completed" ? "#78c6a3" : "#d1a86b"; font.pixelSize: 11; Layout.preferredWidth: 115 }
                  Text { text: goalRow.modelData.evidence && goalRow.modelData.evidence.verified ? "Verified" : "Pending"; color: goalRow.modelData.evidence && goalRow.modelData.evidence.verified ? "#78c6a3" : "#969daa"; font.pixelSize: 11; Layout.preferredWidth: 100 }
                }
              }
              footer: Text {
                visible: root.goals.length === 0
                text: root.apiOnline ? "No goals yet. Run the local evidence check to verify the goal pipeline." : "Waiting for the local Core service."
                color: "#969daa"
                font.pixelSize: 12
                padding: 12
              }
            }
          }
        }

        RowLayout {
          Layout.fillWidth: true
          Rectangle { Layout.fillWidth: true; Layout.preferredHeight: 1; color: "#2a2e37" }
          Button {
            id: stopIntimacyButton
            text: "Stop intimate processing"
            enabled: root.apiOnline && !root.requestActive
            onClicked: root.stopIntimacy()
            contentItem: Text { text: stopIntimacyButton.text; color: "#f1eee7"; font.pixelSize: 11; horizontalAlignment: Text.AlignHCenter; verticalAlignment: Text.AlignVCenter }
            background: Rectangle { radius: 8; color: stopIntimacyButton.down ? "#552d34" : "#3a252b"; border.color: "#74404a" }
            Layout.preferredHeight: 34
            Layout.preferredWidth: 180
          }
        }
        Text {
          text: "This build runs only a public-fixture checksum. Camera, microphone, screen capture, local inference and general planning are not connected."
          color: "#777f8c"
          font.pixelSize: 10
          wrapMode: Text.WordWrap
          Layout.fillWidth: true
        }
      }
    }
  }
}
