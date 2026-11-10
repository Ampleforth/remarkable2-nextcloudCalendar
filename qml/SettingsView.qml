// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick

Rectangle {
    id:page
    required property var app
    color:"#f4f4f4"
    MouseArea {anchors.fill:parent}
    EdgeBack {objectName:"settingsBackGesture";onBack:app.closeSettings()}
    Item {id:heading;x:32;y:24;width:parent.width-64;height:64
        Text {text:"Settings";font.pixelSize:36;font.bold:true;color:"black"}
        Action {anchors.right:parent.right;label:"Done";onTriggered:app.closeSettings()}
    }
    Flickable {
        anchors.top:heading.bottom;anchors.topMargin:20;anchors.bottom:parent.bottom;anchors.bottomMargin:app.keyboardHeight;clip:true;width:parent.width;contentHeight:form.height+48
        Column {
            id:form;x:32;width:parent.width-64;spacing:18
            Text {text:"Nextcloud connection";font.pixelSize:28;font.bold:true;color:"black"}
            Field {objectName:"serverField";width:parent.width;label:"Server URL";value:app.connectionDraft.server;editable:!app.status.busy&&!app.connectionSaving;onActivated:app.field="connection.server"}
            Field {width:parent.width;label:"Username";value:app.connectionDraft.username;editable:!app.status.busy&&!app.connectionSaving;onActivated:app.field="connection.username"}
            Field {objectName:"passwordField";width:parent.width;label:"App password";value:app.connectionDraft.app_password?"••••••••":(app.connectionDraft.has_password?"Saved · leave blank to keep it":"Tap to enter app password");editable:!app.status.busy&&!app.connectionSaving;onActivated:app.field="connection.app_password"}
            Action {width:230;label:app.connectionSaving?"Saving…":"Save & sync";enabled:!app.status.busy&&!app.connectionSaving;onTriggered:app.saveConnection()}
            Text {width:parent.width;text:app.message||app.status.message;font.pixelSize:21;wrapMode:Text.Wrap;color:"#444444"}
            Text {text:"Time format";font.pixelSize:28;font.bold:true;color:"black"}
            Row {spacing:12
                Action {width:170;label:"12-hour";selected:app.timeFormat==="12h";onTriggered:app.savePreference("time_format","12h")}
                Action {width:170;label:"24-hour";selected:app.timeFormat==="24h";onTriggered:app.savePreference("time_format","24h")}
            }
            Text {width:parent.width;text:app.status.last_sync?"Last sync: "+app.status.last_sync:"Not synced yet";font.pixelSize:20;color:"#555555"}
            Text {text:"Visible calendars & task lists";font.pixelSize:28;font.bold:true;color:"black"}
            Repeater {model:app.status.collections;delegate:Action {
                required property var modelData
                width:form.width;checkable:true;checked:modelData.enabled;label:modelData.name+(modelData.writable?"":" · read-only");enabled:!app.status.busy
                onTriggered:app.request("POST","/collections",{id:modelData.id,enabled:!modelData.enabled},function(){app.refresh()})
            }}
            Text {text:"Pending changes: "+app.status.pending.length;font.pixelSize:28;font.bold:true;color:"black"}
            Repeater {model:app.status.pending;delegate:Column {
                required property var modelData
                property bool discardConfirm:false
                width:form.width;spacing:10
                Text {width:parent.width;text:(modelData.title||"Item")+" · "+(modelData.delete?"Deletion":"Edit")+"\n"+(modelData.error||"Waiting to sync");font.pixelSize:22;wrapMode:Text.Wrap;color:"black"}
                Row {spacing:12
                    Action {label:"Retry";enabled:!app.status.busy&&modelData.error.indexOf("conflict:")!==0;onTriggered:app.request("POST","/pending",{id:modelData.id,action:"retry"},function(){app.sync()})}
                    Action {width:260;label:parent.parent.discardConfirm?"Confirm discard":"Discard local change";enabled:!app.status.busy;onTriggered:{if(parent.parent.discardConfirm)app.request("POST","/pending",{id:modelData.id,action:"discard"},function(){app.refresh()});else parent.parent.discardConfirm=true}}
                }
            }}
        }
    }
}
