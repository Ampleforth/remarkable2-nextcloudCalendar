// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

CalendarGesture {
    signal back
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.bottom: parent.bottom
    width: 64
    z: 100
    minimumSwipe: 100
    // Keep the grab when moving left beyond the narrow edge strip.
    preventStealing: true
    propagateComposedEvents: true
    onSwiped: function(direction) { if (direction === 1) back() }
    onClicked: function(mouse) { mouse.accepted = false }
}
