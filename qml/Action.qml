// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

Rectangle {
    id: button
    property string label
    property bool selected: false
    readonly property bool pressed: tap.pressed
    signal triggered
    width: 110; height: 58
    radius: 13
    color: selected ? "black" : "#e9e9e9"
    border.color: selected ? "black" : "#d5d5d5"; border.width: 1
    opacity: enabled ? 1 : 0.4
    Text { anchors.centerIn: parent; width:parent.width-18; horizontalAlignment:Text.AlignHCenter; elide:Text.ElideRight; text: button.label; color: button.selected ? "white" : "black"; font.pixelSize: 22; font.bold:button.selected }
    MouseArea { id:tap; anchors.fill: parent; onClicked: button.triggered() }
}
