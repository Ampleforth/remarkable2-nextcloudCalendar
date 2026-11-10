// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

Rectangle {
    id:page
    required property var app
    readonly property var draft:app.draft
    color:"#f4f4f4"
    MouseArea {anchors.fill:parent}
    EdgeBack {objectName:"eventBackGesture";enabled:!app.saving;onBack:app.cancelEdit()}
    Column {
        id:heading;x:24;y:18;width:parent.width-48;spacing:14
        Row {spacing:14
            Text {width:Math.max(200,heading.width-390);text:page.draft?(page.draft.kind==="VTODO"?"Task":"Event"):"";font.pixelSize:36;font.bold:true;color:"black"}
            Action {label:"Cancel";onTriggered:app.cancelEdit()}
            Action {label:"Save";selected:true;enabled:page.draft&&page.draft.writable&&!app.status.busy&&!app.saving;onTriggered:app.save(false)}
            Action {label:app.confirmingDelete?"Confirm":"Delete";enabled:page.draft&&page.draft.id!==""&&page.draft.writable&&!app.status.busy&&!app.saving;onTriggered:{if(app.confirmingDelete)app.save(true);else app.confirmingDelete=true}}
        }
        Text {width:parent.width;text:app.message||(page.draft&&!page.draft.writable?"Read-only · edit recurring items and invitations in Nextcloud":"Saved on the tablet until you tap Sync.");font.pixelSize:20;wrapMode:Text.Wrap;color:"#444444"}
        Text {text:"Swipe left from the right edge to go back";font.pixelSize:18;color:"#555555"}
    }
    Flickable {
        anchors.top:heading.bottom;anchors.topMargin:20;anchors.left:parent.left;anchors.right:parent.right;anchors.bottom:parent.bottom;anchors.bottomMargin:app.keyboardHeight;clip:true;contentHeight:form.height+48
        Column {
            id:form;x:24;width:parent.width-48;spacing:18
            Repeater {model:page.draft?["title","description","location"]:[];delegate:Field {
                required property string modelData
                width:form.width;label:modelData.charAt(0).toUpperCase()+modelData.slice(1);value:page.draft?page.draft[modelData]:"";editable:page.draft&&page.draft.writable;onActivated:app.field=modelData
            }}
            Field {id:calendar;objectName:"calendarDropdown";width:form.width;label:"Calendar / task list";value:page.draft?page.draft.calendar_name:"";dropdown:page.draft&&page.draft.id==="";editable:page.draft&&page.draft.id==="";onActivated:app.chooseCalendar(calendar)}
            Action {visible:page.draft&&page.draft.kind==="VEVENT";width:220;checkable:true;checked:page.draft&&page.draft.all_day;label:"All day";enabled:page.draft&&page.draft.writable;onTriggered:{var all=!page.draft.all_day;app.setField("all_day",all);app.setField("start",all?page.draft.start.slice(0,10):page.draft.start.slice(0,10)+"T09:00");app.setField("end",all?app.iso(app.add(app.date(page.draft.start),1)):page.draft.start.slice(0,10)+"T10:00")}}
            Repeater {model:page.draft&&page.draft.kind==="VEVENT"?["start","end"]:[];delegate:Row {
                required property string modelData
                width:form.width;spacing:16
                Field {width:page.draft&&page.draft.all_day?form.width:(form.width-16)/2;label:(modelData==="start"?"Start date":"End date")+(page.draft&&page.draft.all_day&&modelData==="end"?" (exclusive)":"")+" · YYYY-MM-DD";value:app.dateValue(modelData);editable:page.draft&&page.draft.writable;onActivated:app.field="date."+modelData}
                Field {id:time;objectName:modelData+"TimeDropdown";visible:page.draft&&!page.draft.all_day;width:(form.width-16)/2;label:modelData==="start"?"Start time":"End time";value:page.draft?app.formatTime(page.draft[modelData].slice(11,16)):"";dropdown:true;editable:page.draft&&page.draft.writable;onActivated:app.chooseTime(time,modelData)}
            }}
            Repeater {model:page.draft&&page.draft.kind==="VTODO"?["scheduled","due"]:[];delegate:Field {
                required property string modelData
                width:form.width;label:(modelData==="scheduled"?"Scheduled for":"Due by")+" · YYYY-MM-DD or YYYY-MM-DDTHH:MM (24h)";value:page.draft?page.draft[modelData]||"":"";editable:page.draft&&page.draft.writable;onActivated:app.field=modelData
            }}
            Field {id:priority;visible:page.draft&&page.draft.kind==="VTODO";width:form.width;label:"Priority";value:page.draft?app.priorityLabel(page.draft.priority):"None";dropdown:true;editable:page.draft&&page.draft.writable;onActivated:app.choosePriority(priority)}
            Action {visible:page.draft&&page.draft.kind==="VTODO";width:240;checkable:true;checked:!!(page.draft&&page.draft.completed);label:"Completed";enabled:page.draft&&page.draft.writable;onTriggered:app.setField("completed",!page.draft.completed)}
        }
    }
}
