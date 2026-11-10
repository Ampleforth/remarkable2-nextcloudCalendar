// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

MouseArea {
    id: gesture
    signal swiped(int direction)
    signal tapped(real tapX, real tapY)
    signal held(real holdX, real holdY)
    property real startX: 0
    property real startY: 0
    property bool tapCandidate: false
    property bool horizontal: false
    property bool wasHeld: false
    property real minimumSwipe: Math.max(80, Math.min(160, width * 0.10))
    acceptedButtons: Qt.LeftButton
    preventStealing: false
    onPressed: function(mouse) {
        startX = mouse.x; startY = mouse.y
        horizontal = true; tapCandidate = true; wasHeld = false
    }
    onPositionChanged: function(mouse) {
        var dx = mouse.x - startX, dy = mouse.y - startY
        if (Math.abs(dx) > 12 || Math.abs(dy) > 12) tapCandidate = false
        if (Math.abs(dy) > 24 && Math.abs(dy) >= Math.abs(dx)) horizontal = false
    }
    onReleased: function(mouse) {
        var dx = mouse.x - startX, dy = mouse.y - startY
        if (!wasHeld && horizontal && Math.abs(dx) >= minimumSwipe && Math.abs(dx) > 2 * Math.abs(dy))
            swiped(dx < 0 ? 1 : -1)
        else if (!wasHeld && tapCandidate && Math.abs(dx) <= 12 && Math.abs(dy) <= 12)
            tapped(mouse.x, mouse.y)
        tapCandidate = false; horizontal = false
    }
    onPressAndHold: function(mouse) {
        if (tapCandidate) { wasHeld = true; held(mouse.x, mouse.y) }
    }
    onCanceled: { tapCandidate = false; horizontal = false; wasHeld = false }
}
