// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

Column {
    id: field
    property string label
    property string value
    property bool editable: true
    property bool dropdown: false
    signal activated
    spacing: 5
    Text { text: field.label; font.pixelSize: 19; font.bold: true; color: "#555555" }
    Rectangle {
        width: field.width; height: Math.max(58, content.implicitHeight + 20)
        radius:12;color: "white"; border.color: "#bbbbbb"
        Text { id: content; x: 14; y: 10; width: parent.width - (field.dropdown?64:28); text: field.value || "—"; font.pixelSize: 26; color: "black"; wrapMode: Text.Wrap }
        Text {visible:field.dropdown;anchors.right:parent.right;anchors.rightMargin:18;anchors.verticalCenter:parent.verticalCenter;text:"▾";font.pixelSize:24;color:"black"}
        MouseArea { anchors.fill: parent; enabled: field.editable; onClicked: field.activated() }
    }
}
