// SPDX-License-Identifier: GPL-3.0-only
pragma Singleton
import QtQuick

QtObject {
    property bool colorDisplay: false
    property var calendarColors: ({})
    readonly property color accent: colorDisplay?"#0082c9":"black"
    readonly property var choices: [
        {value:"#0082c9",label:"Blue"},{value:"#3a9d5d",label:"Green"},
        {value:"#e7a52b",label:"Amber"},{value:"#7952b3",label:"Purple"},
        {value:"#167d8d",label:"Teal"},{value:"#c94758",label:"Red"},
        {value:"#bf6a22",label:"Orange"},{value:"#667085",label:"Slate"}
    ]
    function calendarColor(id) {
        var chosen=calendarColors[id]
        if(chosen&&/^#[0-9a-fA-F]{6}$/.test(chosen))return chosen
        var hash=0
        for(var i=0;i<id.length;i++)hash=(hash*31+id.charCodeAt(i))>>>0
        return choices[hash%choices.length].value
    }
    function colorLabel(id) {
        var value=calendarColor(id),found=choices.find(function(c){return c.value===value})
        return found?found.label:value
    }
    function tint(color) {
        var c=Qt.color(color)
        return Qt.rgba(0.82+c.r*0.18,0.82+c.g*0.18,0.82+c.b*0.18,1)
    }
    function calendarFill(id,gray) {return colorDisplay?tint(calendarColor(id)):gray}
    function calendarEdge(id,gray) {return colorDisplay?calendarColor(id):gray}
}
