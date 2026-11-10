// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick
import "Timeline.js" as Geometry

Item {
    id: timeline
    required property var app
    property int dayCount: app.view === "week" ? 7 : 1
    property date firstDay: dayCount === 7 ? app.monday(app.selected) : app.selected
    property real hourHeight: 112
    readonly property int startHour: 5
    readonly property int endHour: 22
    readonly property real gutter: 88
    readonly property real dayWidth: Math.max(1, (width - gutter) / dayCount)
    readonly property var days: {
        var result = []
        for (var i = 0; i < dayCount; ++i) {
            var day = app.add(firstDay, i)
            result.push({date:day, key:app.iso(day), entries:Geometry.layout(app.items, app.iso(day), startHour * 60, endHour * 60),
                allDay:app.eventsOn(day).filter(function(e){return !!e.all_day})})
        }
        return result
    }
    readonly property real allDayHeight: Math.max(42, days.reduce(function(max, day) {
        return Math.max(max, day.allDay.length * 48)
    }, 0))
    property alias scroll: hours
    property alias gridGesture: gridGesture

    function resetScroll() { hours.contentY = 0 }
    function eventAt(x, y) {
        if (x < gutter || x >= width || y < 0 || y >= (endHour-startHour)*hourHeight) return null
        var index = Math.min(dayCount - 1, Math.floor((x - gutter) / dayWidth))
        return Geometry.hit(days[index].entries, ((x - gutter) / dayWidth) - index, startHour * 60 + y * 60 / hourHeight)
    }
    function holdAt(x, y) {
        if (x < gutter || x >= width || y < 0 || y >= (endHour-startHour)*hourHeight || eventAt(x, y)) return
        var index = Math.min(dayCount - 1, Math.floor((x - gutter) / dayWidth))
        var minute = Math.min(endHour*60-30, Math.max(startHour*60, Math.floor((startHour*60+y*60/hourHeight)/30)*30))
        app.newEventAt(days[index].date, minute)
    }
    onFirstDayChanged: Qt.callLater(resetScroll)
    onVisibleChanged: if (visible) Qt.callLater(resetScroll)
    Component.onCompleted: Qt.callLater(resetScroll)

    Row {
        id: headings
        x: timeline.gutter
        height: 72
        Repeater {
            model: timeline.days
            delegate: Rectangle {
                required property var modelData
                width: timeline.dayWidth; height: headings.height
                radius:10;color: modelData.key === app.iso(new Date()) ? "black" : "#eeeeee"
                border.color: "#cccccc"
                Text {
                    anchors.centerIn: parent
                    text: Qt.formatDate(modelData.date, timeline.dayCount === 7 ? "ddd\nd" : "dddd, d MMMM")
                    horizontalAlignment: Text.AlignHCenter
                    font.pixelSize: timeline.dayCount === 7 ? 24 : 28
                    font.bold: true
                    color: modelData.key === app.iso(new Date()) ? "white" : "black"
                }
            }
        }
    }
    CalendarGesture {
        objectName: "timelineHeadingsGesture"
        x: timeline.gutter; width: timeline.width - x; height: headings.height
        enabled: timeline.visible
        onSwiped: function(direction) { app.navigate(direction) }
        onTapped: function(x, y) {
            if (timeline.dayCount === 7) {
                app.openDay(timeline.days[Math.min(6, Math.floor(x / timeline.dayWidth))].date)
            }
        }
    }
    Text {x:8;y:headings.height+10;text:"All day";font.weight:Font.DemiBold;font.pixelSize:18;color:"black"}
    Row {
        id: allDayRow
        x: timeline.gutter; y: headings.height
        height: timeline.allDayHeight
        Repeater {
            model: timeline.days
            delegate: Item {
                id: allDayColumn
                required property var modelData
                width: timeline.dayWidth; height: timeline.allDayHeight
                clip: true
                Rectangle {anchors.fill:parent;color:"#f5f5f5";border.color:"#999999"}
                Column {
                    width: parent.width
                    spacing: 2
                    Repeater {
                        model: allDayColumn.modelData.allDay
                        delegate: Rectangle {
                            required property var modelData
                            x: 3; width: allDayColumn.width - 6; height: 46
                            radius:10;color: "#dddddd"; border.color: "#bbbbbb"
                            Text {anchors.fill:parent;anchors.margins:5;text:modelData.title;font.weight:Font.DemiBold;font.pixelSize:20;color:"black";elide:Text.ElideRight;verticalAlignment:Text.AlignVCenter}
                        }
                    }
                }
                CalendarGesture {
                    objectName: "allDayGesture"
                    anchors.fill: parent
                    onSwiped: function(direction) { app.navigate(direction) }
                    onTapped: function(x, y) {
                        var event = allDayColumn.modelData.allDay[Math.floor(y / 48)]
                        if (event) app.openItem(event)
                    }
                }
            }
        }
    }
    Flickable {
        id: hours
        objectName: "timelineHours"
        anchors.top: allDayRow.bottom
        anchors.bottom: parent.bottom
        width: parent.width
        contentWidth: width
        contentHeight: (timeline.endHour-timeline.startHour) * timeline.hourHeight + 28
        flickableDirection: Flickable.VerticalFlick
        boundsBehavior: Flickable.StopAtBounds
        clip: true
        Repeater {
            model: timeline.endHour-timeline.startHour+1
            delegate: Item {
                required property int index
                width: hours.width; height: index === timeline.endHour-timeline.startHour ? 28 : timeline.hourHeight
                y: index * timeline.hourHeight
                Text {x:8;y:2;width:timeline.gutter-12;text:Geometry.clock((timeline.startHour+index)*60,app.timeFormat);font.weight:Font.DemiBold;font.pixelSize:app.timeFormat==="12h"?16:19;color:"black"}
                Rectangle {x:timeline.gutter;width:parent.width-x;height:1;color:"#999999"}
                Rectangle {visible:index<timeline.endHour-timeline.startHour;x:timeline.gutter;y:timeline.hourHeight/2;width:parent.width-x;height:1;color:"#dddddd"}
            }
        }
        Repeater {
            model: timeline.days
            delegate: Item {
                id: dayColumn
                required property var modelData
                required property int index
                x: timeline.gutter + index * timeline.dayWidth
                width: timeline.dayWidth; height: hours.contentHeight
                Rectangle {width:1;height:parent.height;color:"#999999"}
                Repeater {
                    model: dayColumn.modelData.entries
                    delegate: Rectangle {
                        required property var modelData
                        property var event: modelData.event
                        x: modelData.column * dayColumn.width / modelData.columns + 3
                        width: Math.max(1, dayColumn.width / modelData.columns - 6)
                        y: (modelData.start-timeline.startHour*60) * timeline.hourHeight / 60
                        height: (modelData.displayEnd - modelData.start) * timeline.hourHeight / 60
                        radius:10;color: "#e5e5e5"; border.color: "#888888"; border.width: 1
                        clip: true
                        Rectangle {width:5;height:parent.height;color:"black"}
                        Text {
                            x:10;y:4;width:Math.max(1,parent.width-16);height:parent.height-8
                            text: Geometry.clock(modelData.start,app.timeFormat)+(timeline.dayCount===7&&modelData.columns>1?"":"–"+Geometry.clock(modelData.end,app.timeFormat))+"\n"+event.title+"\n"+event.calendar_name+(event.pending?" · pending":"")
                            font.pixelSize: timeline.dayCount === 7 ? 18 : 24
                            font.bold: true
                            wrapMode: Text.WordWrap
                            elide: Text.ElideRight
                            color: "black"
                        }
                    }
                }
            }
        }
        CalendarGesture {
            id: gridGesture
            objectName: "timelineGridGesture"
            width: hours.width; height: hours.contentHeight
            onSwiped: function(direction) { app.navigate(direction) }
            onTapped: function(x, y) {var event=timeline.eventAt(x,y);if(event)app.openItem(event)}
            onHeld: function(x, y) {timeline.holdAt(x,y)}
        }
    }
}
