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
    Text { text: field.label; font.pixelSize: 19; font.bold: true; color:"black" }
    Rectangle {
        width: field.width; height: Math.max(58, content.implicitHeight + 20)
        radius:13;color: "white"; border.color: "black";border.width:2
        Text { id: content; x: 14; y: 10; width: parent.width - (field.dropdown?64:28); text: field.value || "—"; font.weight:Font.DemiBold;font.pixelSize: 26; color: "black"; wrapMode: Text.Wrap }
        Canvas {visible:field.dropdown;anchors.right:parent.right;anchors.rightMargin:18;anchors.verticalCenter:parent.verticalCenter;width:18;height:12;onPaint:{var c=getContext("2d");c.reset();c.strokeStyle="black";c.lineWidth=2;c.beginPath();c.moveTo(1,2);c.lineTo(9,10);c.lineTo(17,2);c.stroke()}}
        MouseArea { anchors.fill: parent; enabled: field.editable; onClicked: field.activated() }
    }
}
