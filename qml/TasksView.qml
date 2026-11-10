// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

Item {
    id:page
    required property var app
    readonly property var rows:app.tasks()
    Item {
        id:toolbar;x:20;y:18;width:parent.width-40;height:116
        Text {text:"All tasks";font.pixelSize:34;font.bold:true;color:"black"}
        Action {id:sort;dropdown:true;objectName:"taskSortButton";anchors.right:parent.right;width:310;label:"Sort: "+app.sortOptions.filter(function(o){return o.value===app.taskSort})[0].label;onTriggered:app.chooseSort(sort)}
        Row {y:64;spacing:12
            Action {id:lists;dropdown:true;width:300;label:app.taskListName();onTriggered:app.chooseTaskList(lists)}
            Action {width:220;label:app.showCompleted?"Hide completed":"Show completed";onTriggered:app.showCompleted=!app.showCompleted}
        }
    }
    ListView {
        id:list;anchors.top:toolbar.bottom;anchors.topMargin:12;anchors.bottom:parent.bottom;anchors.bottomMargin:16;x:20;width:parent.width-40;clip:true;spacing:10;model:page.rows
        delegate:Column {
            required property var modelData
            required property int index
            property string section:app.taskSection(modelData)
            width:list.width;spacing:10
            Rectangle {
                visible:section!==""&&(index===0||section!==app.taskSection(page.rows[index-1]))
                width:parent.width;height:visible?50:0;radius:12;color:"#e7e7e7"
                Text {x:20;anchors.verticalCenter:parent.verticalCenter;text:section;font.pixelSize:24;font.bold:true;color:"black"}
            }
            Rectangle {
                width:parent.width;height:Math.max(112,title.implicitHeight+metadata.implicitHeight+32);radius:16;color:"white";border.color:"#cccccc"
                CheckMark {
                    x:22;y:30;width:40;height:40;checked:!!modelData.completed
                    MouseArea {anchors.fill:parent;enabled:modelData.writable&&!app.status.busy;onClicked:app.toggleTask(modelData)}
                }
                Text {id:title;x:88;y:18;width:parent.width-250;text:modelData.title;font.pixelSize:28;font.bold:true;font.strikeout:!!modelData.completed;wrapMode:Text.Wrap;color:"black"}
                Text {id:metadata;x:88;anchors.top:title.bottom;anchors.topMargin:6;width:parent.width-120;text:app.taskSubtitle(modelData);font.weight:Font.DemiBold;font.pixelSize:20;wrapMode:Text.Wrap;color:"black"}
                Rectangle {visible:modelData.priority>0;anchors.right:parent.right;anchors.rightMargin:20;y:20;width:110;height:38;radius:10;color:"#dddddd";Text{anchors.centerIn:parent;text:app.priorityLabel(modelData.priority);font.pixelSize:20;font.bold:true;color:"black"}}
                MouseArea {x:78;width:parent.width-78;height:parent.height;onClicked:app.openItem(modelData)}
            }
        }
        Text {visible:parent.count===0;anchors.centerIn:parent;text:"No tasks in this view";font.weight:Font.DemiBold;font.pixelSize:27;color:"black"}
    }
}
