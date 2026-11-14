// SPDX-License-Identifier: GPL-3.0-only
// Fictional read-only demo. All requests are overridden; no server is contacted.
import QtQuick
import QtQuick.Window
import "../qml" as Calendar

Window {
    id:window
    width:1404;height:1872;visible:true
    property string demoView: "week"
    property string capturePath: ""
    property bool keyboardPreview: false
    property bool colorPreview: false
    property bool settingsPreview: false
    property bool importPreview: false
    Calendar.Main {
        id:app;anchors.fill:parent;view:window.demoView;selected:new Date(2026,10,10,12)
        function request(method,path,data,done) {
            if(path==="/status")done({color_display:window.colorPreview,calendar_colors:{work:"#0082c9",personal:"#3a9d5d",projects:"#7952b3"},busy:false,message:"Demo · fictional data",last_sync:"",pending:[],time_format:"12h",task_sort:"scheduled",collections:[
                {id:"personal",name:"Personal",events:true,tasks:true,enabled:true,writable:true},
                {id:"work",name:"Work",events:true,tasks:true,enabled:true,writable:true},
                {id:"projects",name:"Projects",events:true,tasks:true,enabled:true,writable:true}
            ]})
            else if(path==="/captures")done({captures:[{id:"demo-capture-1",summary:"Book equipment for next week",description:"Captured from Project notebook · p.3"},{id:"demo-capture-2",summary:"Review release checklist",description:"Captured from Project notebook · p.4"},{id:"demo-capture-3",summary:"Call the dentist",description:"Captured from Personal notebook · p.2"}]})
            else if(path==="/launcher")done({enabled:true})
            else if(path==="/connection")done({server:"https://cloud.example.com",username:"demo-user",has_password:false})
            else if(path.indexOf("/items?")===0)done({items:[
                event("Planning","2026-11-09T09:00","2026-11-09T10:00","Work"),
                event("Project review","2026-11-10T10:00","2026-11-10T11:00","Work"),
                event("Design session","2026-11-10T10:30","2026-11-10T11:30","Projects"),
                event("Dentist","2026-11-10T14:00","2026-11-10T15:00","Personal"),
                event("Equipment delivery","2026-11-12T00:00","2026-11-13T00:00","Personal",true),
                event("Team offsite","2026-11-12T00:00","2026-11-13T00:00","Work",true),
                event("Release preparation","2026-11-12T00:00","2026-11-13T00:00","Projects",true),
                event("Dinner","2026-11-13T19:00","2026-11-13T20:30","Personal"),
                task("Review project notes","2026-11-10T09:00","2026-11-12",1,"Work"),
                task("Call the dentist","2026-11-10T11:00","",5,"Personal"),
                task("Review design draft","2026-11-10T14:00","",1,"Projects"),
                task("Water the plants","2026-11-11","",5,"Personal"),
                task("Plan weekend trip","","",0,"Personal")
            ]})
            else if(path==="/preferences") {var next=JSON.parse(JSON.stringify(status));Object.keys(data).forEach(function(key){next[key]=data[key]});done(next)}
            else done({ok:true})
        }
        function event(title,start,end,calendar,allDay) {return {id:title,kind:"VEVENT",title:title,start:start,end:end,calendar:calendar.toLowerCase(),calendar_name:calendar,all_day:!!allDay,writable:true,pending:false,completed:false,description:"",location:"",revision:"demo",scheduled:"",priority:0,due:""}}
        function task(title,start,due,priority,calendar) {return {id:title,kind:"VTODO",title:title,scheduled:start,start:start,due:due,priority:priority,calendar:calendar.toLowerCase(),calendar_name:calendar,writable:true,completed:false,pending:false,description:"",location:"",revision:"demo"}}
    }
    Timer {
        interval:400;running:window.capturePath!=="";repeat:false
        onTriggered:{
            if(window.settingsPreview)app.openSettings()
            if(window.importPreview)app.openCapture()
            if(window.keyboardPreview){app.newItem();app.setField("title","Team planning")}
            app.grabToImage(function(result){if(!result.saveToFile(window.capturePath))Qt.exit(1);else Qt.quit()})
        }
    }
}
