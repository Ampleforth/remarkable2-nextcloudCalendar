// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick
import net.asivery.AppLoad 1.0
import "Timeline.js" as Geometry

Rectangle {
    id: root
    anchors.fill: parent
    color: "#f4f4f4"
    signal close
    AppLoad { id:endpoint; applicationID:"org.remarkable.nextcloudcalendar" }
    property string view: "week"
    property date selected: new Date()
    property var items: []
    property var status: ({busy:false,collections:[],pending:[],message:"Tap Sync to connect to Nextcloud"})
    property string message: ""
    property bool settingsOpen: false
    property var connectionDraft: ({server:"https://cloud.example.com",username:"",app_password:"",has_password:false})
    property bool connectionSaving: false
    property var draft: null
    property string field: ""
    function hideNativeKeyboard() {
        if(Qt.inputMethod)Qt.inputMethod.hide()
        var ancestor=root.parent
        for(var i=0;ancestor&&i<20;i++,ancestor=ancestor.parent) {
            if(typeof ancestor.virtualKeyboardRef!=="undefined"&&ancestor.virtualKeyboardRef)
                ancestor.virtualKeyboardRef.active=false
        }
        root.forceActiveFocus()
    }
    onFieldChanged:if(field!=="")hideNativeKeyboard()
    property var dateFields: ({})
    property bool showCompleted: false
    property string taskList: ""
    property bool confirmingDelete: false
    property bool saving: false
    property bool loading: false
    property var requests: []
    property int generation: 0
    readonly property string timeFormat: status.time_format || "24h"
    readonly property string taskSort: status.task_sort || "scheduled"
    readonly property var weekdays: ["Mon","Tue","Wed","Thu","Fri","Sat","Sun"]
    readonly property var sortOptions: [{value:"scheduled",label:"Scheduled date"},{value:"due",label:"Due date"},{value:"priority",label:"Priority"},{value:"title",label:"Title"}]
    readonly property var priorityOptions: [{value:0,label:"None"},{value:1,label:"High"},{value:5,label:"Normal"},{value:9,label:"Low"}]
    function unloading() {poll.stop();requests.forEach(function(x){x.abort()});endpoint.terminate()}
    function iso(d) {return d.getFullYear()+"-"+String(d.getMonth()+1).padStart(2,"0")+"-"+String(d.getDate()).padStart(2,"0")}
    function date(s) {return new Date(Number(s.slice(0,4)),Number(s.slice(5,7))-1,Number(s.slice(8,10)),12)}
    function add(d,n) {var t=new Date(d);t.setDate(t.getDate()+n);return t}
    function monday(d) {return add(d,-((d.getDay()+6)%7))}
    function monthStart() {return monday(new Date(selected.getFullYear(),selected.getMonth(),1,12))}
    function rangeStart() {return view==="month"?monthStart():(view==="day"?selected:monday(selected))}
    function periodTitle() {
        if(view==="month")return Qt.formatDate(selected,"MMMM yyyy")
        if(view!=="week")return Qt.formatDate(selected,"dddd, d MMM yyyy")
        var first=monday(selected),last=add(first,6)
        if(first.getFullYear()!==last.getFullYear())return Qt.formatDate(first,"MMM d, yyyy")+" – "+Qt.formatDate(last,"MMM d, yyyy")
        if(first.getMonth()!==last.getMonth())return Qt.formatDate(first,"MMM d")+" – "+Qt.formatDate(last,"MMM d, yyyy")
        return Qt.formatDate(first,"MMM d")+"–"+last.getDate()+", "+last.getFullYear()
    }
    function formatTime(value) {return Geometry.clock(Number(value.slice(0,2))*60+Number(value.slice(3,5)),timeFormat)}
    function displayDate(value) {if(!value)return "";return Qt.formatDate(date(value),"d MMM yyyy")+(value.length>10?" · "+formatTime(value.slice(11,16)):"")}
    function request(method,path,data,done) {
        var xhr=new XMLHttpRequest();requests=requests.concat([xhr])
        xhr.onreadystatechange=function(){
            if(xhr.readyState!==XMLHttpRequest.DONE)return
            requests=requests.filter(function(x){return x!==xhr})
            try {var value=JSON.parse(xhr.responseText);if(xhr.status!==200||value.error)throw value.error||"Backend unavailable";done(value)}
            catch(e){message=String(e);saving=false;connectionSaving=false;loading=false}
        }
        xhr.open(method,"http://127.0.0.1:38766"+path)
        if(method!=="GET"){xhr.setRequestHeader("Content-Type","application/json");xhr.setRequestHeader("X-RemarkableCalendar-Request","sync")}
        xhr.send(data===null?null:JSON.stringify(data))
    }
    function refresh() {
        var serial=++generation,start=rangeStart(),end=add(start,view==="month"?42:(view==="day"?1:7));loading=true
        request("GET","/items?start="+iso(start)+"&end="+iso(end),null,function(v){if(serial===generation){items=v.items;loading=false}})
        request("GET","/status",null,function(v){status=v})
    }
    function sync() {request("POST","/sync",{},function(){message="";refresh()})}
    function navigate(n) {selected=view==="month"?new Date(selected.getFullYear(),selected.getMonth()+n,1,12):add(selected,n*(view==="week"?7:1));refresh()}
    function changeView(v) {view=v;refresh()}
    function openDay(day) {selected=day;changeView("day")}
    function cancelEdit() {draft=null;field="";dateFields={};confirmingDelete=false;picker.close()}
    function eventsOn(d) {
        var key=iso(d),next=iso(add(d,1))+"T00:00"
        return items.filter(function(i){return i.kind==="VEVENT"&&i.start<next&&(i.end>key+"T00:00"||i.start.slice(0,10)===key)})
    }
    function taskDate(item,kind) {return kind==="due"?(item.due||""):(item.scheduled||item.start||"")}
    function tasks() {
        return items.filter(function(i){return i.kind==="VTODO"&&(showCompleted||!i.completed)&&(!taskList||i.calendar===taskList)}).sort(function(a,b){
            if(taskSort==="priority") {
                var pa=a.priority>0?a.priority:10,pb=b.priority>0?b.priority:10
                if(pa!==pb)return pa-pb
            } else if(taskSort!=="title") {
                var first=taskDate(a,taskSort),second=taskDate(b,taskSort)
                if(!first&&second)return 1
                if(first&&!second)return -1
                if(first<second)return -1
                if(first>second)return 1
            }
            return a.title.localeCompare(b.title)
        })
    }
    function priorityLabel(value) {return value>0?(value<=4?"High":(value===5?"Normal":"Low")):"None"}
    function taskSection(item) {
        if(taskSort==="title")return ""
        if(taskSort==="priority")return priorityLabel(item.priority)+" priority"
        var value=taskDate(item,taskSort)
        if(!value)return taskSort==="due"?"No due date":"Unscheduled"
        var key=value.slice(0,10),today=new Date()
        if(key===iso(today))return "Today · "+Qt.formatDate(today,"ddd d")
        if(key===iso(add(today,1)))return "Tomorrow · "+Qt.formatDate(date(key),"ddd d")
        return Qt.formatDate(date(key),"dddd, d MMM yyyy")
    }
    function taskSubtitle(item) {
        var scheduled=taskDate(item,"scheduled")
        return (scheduled?displayDate(scheduled):"Unscheduled")+" · "+item.calendar_name+(item.due?" · Due "+displayDate(item.due):"")+(item.pending?" · pending":"")
    }
    function savePreference(name,value) {var p={};p[name]=value;request("POST","/preferences",p,function(v){status=v})}
    function chooseSort(anchor) {field="";picker.show(anchor,"Sort tasks",sortOptions,taskSort,function(value){savePreference("task_sort",value)})}
    function chooseTaskList(anchor) {
        var options=[{value:"",label:"All lists"}].concat(status.collections.filter(function(c){return c.tasks}).map(function(c){return {value:c.id,label:c.name}}))
        picker.show(anchor,"Task list",options,taskList,function(value){taskList=value})
    }
    function taskListName() {var found=status.collections.find(function(c){return c.id===taskList});return found?found.name:"All lists"}
    function openItem(item) {dateFields={};draft=JSON.parse(JSON.stringify(item));if(draft.all_day){setField("start",draft.start.slice(0,10));setField("end",draft.end.slice(0,10))};field="";confirmingDelete=false;message=item.error||""}
    function newItem() {
        dateFields={}
        var kind=view==="tasks"?"VTODO":"VEVENT"
        var calendars=status.collections.filter(function(c){return c.writable&&(kind==="VTODO"?c.tasks:c.events)})
        if(!calendars.length){message="Sync first, then choose a writable calendar or task list.";return}
        draft={id:"",calendar:calendars[0].id,calendar_name:calendars[0].name,kind:kind,title:"",description:"",location:"",start:kind==="VTODO"?"":iso(selected)+"T09:00",end:kind==="VTODO"?"":iso(selected)+"T10:00",scheduled:"",priority:0,due:"",all_day:false,completed:false,recurring:false,writable:true,pending:false,error:"",revision:""}
        setField("create_id","xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx".replace(/x/g,function(){return Math.floor(Math.random()*16).toString(16)}))
        field="title";confirmingDelete=false;message=""
    }
    function newEventAt(day,minute) {
        if(status.busy)return
        selected=day;newItem();if(!draft)return
        setField("start",iso(day)+"T"+Geometry.clock(minute))
        setField("end",iso(minute+60>=1440?add(day,1):day)+"T"+Geometry.clock((minute+60)%1440))
    }
    function setField(name,value) {
        if(!draft)return
        if((name==="start"||name==="end")&&dateFields[name]!==undefined&&dateFields[name]===(draft[name]||"").slice(0,10)) {
            var dates=JSON.parse(JSON.stringify(dateFields));dates[name]=String(value).slice(0,10);dateFields=dates
        }
        var d=JSON.parse(JSON.stringify(draft));d[name]=value;draft=d
    }
    function dateValue(name) {return dateFields[name]!==undefined?dateFields[name]:(draft?(draft[name]||"").slice(0,10):"")}
    function validDate(value) {
        if(!/^\d{4}-\d{2}-\d{2}$/.test(value))return false
        var y=Number(value.slice(0,4)),m=Number(value.slice(5,7)),day=Number(value.slice(8,10)),d=new Date(0)
        d.setFullYear(y,m-1,day);d.setHours(12,0,0,0)
        return y>=1&&y<=9998&&d.getFullYear()===y&&d.getMonth()===m-1&&d.getDate()===day
    }
    function editValue() {
        if(field.indexOf("connection.")===0)return connectionDraft[field.slice(11)]||""
        if(!draft)return ""
        if(field.indexOf("date.")===0)return dateValue(field.slice(5))
        return draft[field]||""
    }
    function editText(value) {
        if(field.indexOf("connection.")===0){var c=JSON.parse(JSON.stringify(connectionDraft));c[field.slice(11)]=value;connectionDraft=c}
        else if(field.indexOf("date.")===0){
            var key=field.slice(5),dates=JSON.parse(JSON.stringify(dateFields));dates[key]=value;dateFields=dates
            if(validDate(value))setField(key,value+(draft.all_day?"":"T"+(draft[key]||"2000-01-01T09:00").slice(11,16)))
        }
        else setField(field,value)
    }
    function chooseTime(anchor,key) {
        field=""
        var options=[]
        for(var minute=0;minute<1440;minute+=15){var value=Geometry.clock(minute);options.push({value:value,label:formatTime(value)})}
        picker.show(anchor,key==="start"?"Start time":"End time",options,(draft[key]||"").slice(11,16),function(value){
            setField(key,(draft[key]||iso(selected)).slice(0,10)+"T"+value)
            if(key==="start"&&draft.end<=draft.start){var minute=Number(value.slice(0,2))*60+Number(value.slice(3))+60;setField("end",iso(minute>=1440?add(date(draft.start),1):date(draft.start))+"T"+Geometry.clock(minute%1440))}
        })
    }
    function chooseCalendar(anchor) {
        field=""
        var options=status.collections.filter(function(c){return c.writable&&(draft.kind==="VTODO"?c.tasks:c.events)}).map(function(c){return {value:c.id,label:c.name}})
        picker.show(anchor,"Destination calendar",options,draft.calendar,function(value){var c=status.collections.find(function(c){return c.id===value});setField("calendar",value);setField("calendar_name",c.name)})
    }
    function choosePriority(anchor) {field="";picker.show(anchor,"Priority",priorityOptions,draft.priority||0,function(value){setField("priority",value)})}
    function save(remove) {
        if(saving)return
        if(!remove&&draft.kind==="VEVENT"&&(!validDate(dateValue("start"))||!validDate(dateValue("end")))){message="Enter valid start and end dates as YYYY-MM-DD";return}
        saving=true;var payload=JSON.parse(JSON.stringify(draft));payload.delete=remove
        request("POST","/items",payload,function(){saving=false;cancelEdit();refresh();message="Saved on tablet · tap Sync to upload"})
    }
    function toggleTask(i) {if(!i.writable||status.busy)return;var e=JSON.parse(JSON.stringify(i));e.completed=!e.completed;request("POST","/items",e,function(){refresh()})}
    function openSettings() {
        settingsOpen=true;field="";message=""
        request("GET","/connection",null,function(c){if(settingsOpen)connectionDraft={server:c.server,username:c.username,app_password:"",has_password:c.has_password}})
    }
    function closeSettings() {settingsOpen=false;field="";connectionDraft.app_password="";picker.close()}
    function saveConnection() {
        if(connectionSaving)return
        connectionSaving=true;field=""
        request("POST","/connection",{server:connectionDraft.server,username:connectionDraft.username,app_password:connectionDraft.app_password},function(c){
            connectionSaving=false;connectionDraft={server:c.server,username:c.username,app_password:"",has_password:c.has_password};message="Connection saved";sync()
        })
    }
    Timer {id:poll;interval:2000;running:true;repeat:true;onTriggered:root.request("GET","/status",null,function(v){var wasBusy=root.status.busy;root.status=v;if(wasBusy&&!v.busy)root.refresh()})}
    Timer {interval:8000;running:true;repeat:true;onTriggered:{if(root.loading){root.requests.forEach(function(x){x.abort()});root.requests=[];root.loading=false;root.message="Backend request timed out · tap a view to retry"}}}
    Component.onCompleted:refresh()

    Column {
        id:header;x:32;y:24;width:parent.width-64;spacing:18
        Row {
            width:parent.width;spacing:16
            Image {source:"calendar.svg";width:64;height:64;sourceSize.width:128;sourceSize.height:128}
            Text {width:Math.max(200,header.width-400);height:64;verticalAlignment:Text.AlignVCenter;text:"Nextcloud Calendar";font.pixelSize:34;font.bold:true;color:"black"}
            ToolbarButton {icon:"add";enabled:!root.status.busy;onTriggered:root.newItem()}
            ToolbarButton {icon:"sync";busy:root.status.busy;enabled:!root.status.busy;onTriggered:root.sync()}
            ToolbarButton {icon:"settings";onTriggered:root.openSettings()}
            ToolbarButton {icon:"close";onTriggered:root.close()}
        }
        Item {
            width:parent.width;height:62
            Rectangle {width:350;height:62;radius:16;color:"#e8e8e8"
                Row {x:5;y:2;spacing:5
                    Repeater {model:["week","month","day"];delegate:Action {required property string modelData;width:110;label:modelData.charAt(0).toUpperCase()+modelData.slice(1);selected:root.view===modelData;onTriggered:root.changeView(modelData)}}
                }
            }
            Action {objectName:"tasksNavigation";anchors.right:parent.right;width:145;label:"Tasks";selected:root.view==="tasks";onTriggered:root.changeView("tasks")}
        }
        Item {
            visible:root.view!=="tasks";width:parent.width;height:64
            Text {width:parent.width-290;height:64;verticalAlignment:Text.AlignVCenter;text:root.periodTitle();font.pixelSize:34;font.bold:true;color:"black";elide:Text.ElideRight}
            Row {anchors.right:parent.right;spacing:10
                Action {label:"Today";onTriggered:{root.selected=new Date();root.refresh()}}
                ToolbarButton {icon:"previous";onTriggered:root.navigate(-1)}
                ToolbarButton {icon:"next";onTriggered:root.navigate(1)}
            }
        }
        Text {width:parent.width;text:root.message||(root.loading?"Loading…":root.status.message);font.weight:Font.DemiBold;font.pixelSize:18;color:"black";wrapMode:Text.Wrap}
        Text {visible:root.view!=="tasks";text:root.view==="month"?"Swipe to change month · Tap a date to open":"Swipe to browse · Hold an empty time to add an event";font.weight:Font.DemiBold;font.pixelSize:18;color:"black"}
    }
    Rectangle {
        id:content;anchors.top:header.bottom;anchors.topMargin:18;anchors.bottom:parent.bottom;anchors.bottomMargin:24;anchors.left:parent.left;anchors.right:parent.right;anchors.leftMargin:24;anchors.rightMargin:24
        radius:18;color:"white";border.color:"#cccccc";clip:true
        TimelineView {id:timelinePage;objectName:"timelinePage";app:root;anchors.fill:parent;visible:root.view==="week"||root.view==="day"}
        Grid {
            visible:root.view==="month";anchors.fill:parent;columns:7
            Repeater {model:42;delegate:Rectangle {
                id:monthCell;required property int index
                property date day:root.add(root.monthStart(),index)
                width:content.width/7;height:content.height/6;border.color:"#cccccc";radius:6
                color:day.getMonth()===root.selected.getMonth()?"white":"#eeeeee"
                Text {x:12;y:12;text:monthCell.day.getDate();font.pixelSize:28;font.bold:root.iso(monthCell.day)===root.iso(new Date());color:"black"}
                Column {x:8;y:52;width:parent.width-16;spacing:8
                    Repeater {model:root.eventsOn(monthCell.day).slice(0,Math.max(0,Math.floor((monthCell.height-84)/42)));delegate:Text {required property var modelData;width:parent.width;text:modelData.title;font.weight:Font.DemiBold;font.pixelSize:18;maximumLineCount:2;wrapMode:Text.Wrap;elide:Text.ElideRight;color:"black"}}
                }
                Text {anchors.bottom:parent.bottom;anchors.right:parent.right;anchors.margins:8;text:root.eventsOn(monthCell.day).length?root.eventsOn(monthCell.day).length+" events":"";font.weight:Font.DemiBold;font.pixelSize:14;color:"black"}
            }}
        }
        CalendarGesture {
            objectName:"monthGesture";visible:root.view==="month";anchors.fill:parent
            onSwiped:function(direction){root.navigate(direction)}
            onTapped:function(x,y){root.openDay(root.add(root.monthStart(),Math.min(5,Math.floor(y/(height/6)))*7+Math.min(6,Math.floor(x/(width/7)))))}
        }
        TasksView {app:root;anchors.fill:parent;visible:root.view==="tasks"}
    }
    SettingsView {app:root;anchors.fill:parent;visible:root.settingsOpen;z:10}
    EditorView {app:root;anchors.fill:parent;visible:root.draft!==null;z:20}
    TouchKeyboard {
        id:keyboard;objectName:"touchKeyboard";z:40;anchors.bottom:parent.bottom;visible:root.field!==""
        value:root.editValue();label:root.field.indexOf("date.")===0?"Date · YYYY-MM-DD":(root.field==="connection.app_password"?"App password":(root.field==="connection.server"?"Server URL":root.field.replace("connection.","")))
        secret:root.field==="connection.app_password"
        onEdited:function(value){root.editText(value)}
        onDone:root.field=""
    }
    readonly property real keyboardHeight:keyboard.visible?keyboard.height:0
    ChoicePicker {id:picker;objectName:"choicePicker"}
}
