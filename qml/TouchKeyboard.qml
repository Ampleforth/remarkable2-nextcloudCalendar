// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

Rectangle {
    id:keyboard
    property string value: ""
    property string label: ""
    property bool secret: false
    property bool shift: false
    property bool symbols: false
    property bool moreSymbols: false
    readonly property real keyWidth:(width-32-9*8)/10
    signal edited(string value)
    signal done
    width:parent.width;height:rows.height+32;color:"#ededed";border.color:"#aaaaaa"
    function type(value) {
        var position=input.selectionStart
        if(input.selectionEnd>position){input.remove(position,input.selectionEnd);input.insert(position,value)}
        else input.insert(input.cursorPosition,value)
        edited(input.text)
    }
    function erase() {
        if(input.selectionEnd>input.selectionStart)input.remove(input.selectionStart,input.selectionEnd)
        else if(input.cursorPosition>0)input.remove(input.cursorPosition-1,input.cursorPosition)
        edited(input.text)
    }
    component Key:Action {objectName:"key_"+label;height:76;radius:8;color:pressed?"#cccccc":"white";border.color:"#bdbdbd"}
    MouseArea {anchors.fill:parent}
    Column {
        id:rows;x:16;y:16;width:parent.width-32;spacing:10
        Item {
            width:parent.width;height:36
            Text {anchors.verticalCenter:parent.verticalCenter;text:keyboard.label;font.pixelSize:18;color:"#444444"}
            Item {
                objectName:"hideKeyboardButton";anchors.right:parent.right;width:64;height:36
                Canvas {
                    anchors.centerIn:parent;width:40;height:34
                    onPaint:{var c=getContext("2d");c.reset();c.strokeStyle="black";c.lineWidth=2;c.beginPath();c.rect(4,2,32,16);c.moveTo(10,7);c.lineTo(14,7);c.moveTo(18,7);c.lineTo(22,7);c.moveTo(26,7);c.lineTo(30,7);c.moveTo(10,12);c.lineTo(30,12);c.moveTo(13,24);c.lineTo(20,31);c.lineTo(27,24);c.stroke()}
                }
                MouseArea {anchors.fill:parent;onClicked:keyboard.done()}
            }
        }
        Rectangle {width:parent.width;height:64;radius:8;color:"white";border.color:"#bdbdbd"
            TextInput {id:input;objectName:"keyboardInput";x:14;y:10;width:parent.width-28;height:44;font.pixelSize:28;clip:true;selectByMouse:true;text:keyboard.value;echoMode:keyboard.secret?TextInput.Password:TextInput.Normal;onTextEdited:keyboard.edited(text)}
        }
        Repeater {
            model:keyboard.symbols?(keyboard.moreSymbols?["[]{}#%^*+=","_\\|~<>€£¥"]:["1234567890","-/:;()$&@"]):["qwertyuiop","asdfghjkl"]
            delegate:Row {
                required property string modelData
                property string keys:modelData
                x:(rows.width-(keys.length*keyboard.keyWidth+(keys.length-1)*8))/2
                spacing:8
                Repeater {model:parent.keys.length;delegate:Key {required property int index;property string key:parent.keys.charAt(index);width:keyboard.keyWidth;label:keyboard.shift?key.toUpperCase():key;onTriggered:keyboard.type(label)}}
            }
        }
        Row {
            width:parent.width;spacing:8
            Key {width:(rows.width-7*keyboard.keyWidth-8*8)/2;label:keyboard.symbols?(keyboard.moreSymbols?"123":"#+="):(keyboard.shift?"⇧ ON":"⇧");onTriggered:{if(keyboard.symbols)keyboard.moreSymbols=!keyboard.moreSymbols;else keyboard.shift=!keyboard.shift}}
            Repeater {model:keyboard.symbols?[".",",","?","!","'","\"","="]:["z","x","c","v","b","n","m"];delegate:Key {required property string modelData;width:keyboard.keyWidth;label:keyboard.shift?modelData.toUpperCase():modelData;onTriggered:keyboard.type(label)}}
            Key {width:(rows.width-7*keyboard.keyWidth-8*8)/2;label:"⌫";onTriggered:keyboard.erase()}
        }
        Row {width:parent.width;spacing:8
            Key {width:120;label:keyboard.symbols?"ABC":"123";onTriggered:keyboard.symbols=!keyboard.symbols}
            Key {width:76;label:",";onTriggered:keyboard.type(",")}
            Key {width:Math.max(110,rows.width-448);label:"Space";onTriggered:keyboard.type(" ")}
            Key {width:76;label:".";onTriggered:keyboard.type(".")}
            Key {width:144;label:"Done ⏎";onTriggered:keyboard.done()}
        }
    }
}
