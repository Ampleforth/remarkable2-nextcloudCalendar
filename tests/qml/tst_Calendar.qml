// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
import QtQuick
import QtQuick.Window
import QtTest
import "../../qml" as Calendar
import "../../qml/Timeline.js" as Geometry

TestCase {
    id: testCase
    name: "CalendarUI"
    width: 1404
    height: 1872
    visible: true
    when: windowShown
    property var submitted: null
    function initTestCase() {
        testCase.Window.window.width=width
        testCase.Window.window.height=height
        wait(50)
    }
    Calendar.Main {
        id: app
        anchors.fill: parent
        function request(method, path, data, done) {
            if (path === "/status") {
                done({busy:false, message:"Test calendar", last_sync:"", pending:[], collections:[{id:"test", name:"Personal", events:true, tasks:true, enabled:true, writable:true}]})
            } else if (path.indexOf("/items?") === 0) {
                done({items:[{id:"one", calendar:"test", calendar_name:"Personal", kind:"VEVENT", title:"Morning planning", start:iso(selected)+"T09:00", end:iso(selected)+"T10:00", writable:true, all_day:false, pending:false, description:"", location:"", revision:"one"}]})
            } else if(path==="/preferences") {
                testCase.submitted=data
                var next=JSON.parse(JSON.stringify(status));Object.keys(data).forEach(function(key){next[key]=data[key]});done(next)
            } else if(path==="/connection") {
                if(method==="POST")testCase.submitted=data
                done({server:"https://nextcloud.example",username:"test",has_password:true})
            } else {
                testCase.submitted=data
                done({ok:true})
            }
        }
    }
    function init() {
        app.draft=null;app.field="";app.settingsOpen=false;app.view="week";app.taskList="";app.showCompleted=false;app.selected=new Date(2026,10,10,12);app.refresh()
        submitted=null
    }
    function test_week_and_month() {
        compare(app.eventsOn(app.selected).length,1)
        app.changeView("month")
        compare(app.view,"month")
        compare(app.monthStart().getDay(),1)
        app.navigate(1)
        compare(app.selected.getMonth(),11)
        app.changeView("day")
        compare(app.eventsOn(app.selected)[0].title,"Morning planning")
    }
    function test_create_event() {
        app.newItem()
        verify(app.draft!==null)
        compare(app.field,"title")
        app.setField("title","A real event")
        app.setField("start","2026-11-10T11:00")
        app.setField("end","2026-11-10T12:00")
        app.save(false)
        compare(submitted.title,"A real event")
        compare(submitted.kind,"VEVENT")
        compare(app.draft,null)
    }
    function test_all_day_editor() {
        var event=JSON.parse(JSON.stringify(app.items[0]))
        event.all_day=true;event.start="2026-11-10T00:00";event.end="2026-11-11T00:00"
        app.openItem(event)
        compare(app.draft.start,"2026-11-10")
        compare(app.draft.end,"2026-11-11")
    }
    function test_task_create() {
        app.changeView("tasks")
        app.newItem()
        compare(app.draft.kind,"VTODO")
        compare(app.draft.due,"")
        app.setField("title","Water plants")
        app.setField("completed",true)
        app.save(false)
        compare(submitted.completed,true)
    }
    function test_settings_and_keyboard_render() {
        app.settingsOpen=true
        wait(50)
        app.settingsOpen=false
        app.newItem()
        wait(50)
        verify(app.height>1000)
        verify(app.width>1000)
    }
    function test_tasks_merge_lists_in_due_order() {
        app.status={busy:false,pending:[],message:"",task_sort:"due",collections:[
            {id:"work",name:"Work",tasks:true,enabled:true,writable:true},
            {id:"home",name:"Home",tasks:true,enabled:true,writable:true}
        ]}
        app.items=[
            {kind:"VTODO",calendar:"work",title:"Later work task",due:"2026-11-20"},
            {kind:"VTODO",calendar:"home",title:"Undated task",due:""},
            {kind:"VTODO",calendar:"work",title:"Morning task",due:"2026-11-12T09:00"},
            {kind:"VTODO",calendar:"home",title:"Soonest home task",due:"2026-11-10"},
            {kind:"VTODO",calendar:"home",title:"Afternoon task",due:"2026-11-12T14:00"},
            {kind:"VTODO",calendar:"work",title:"Finished task",due:"2026-11-09",completed:true},
            {kind:"VEVENT",calendar:"home",title:"An event",due:"2026-11-08",start:"2026-11-10T09:00",end:"2026-11-10T10:00",calendar_name:"Home"}
        ].map(function(item){item.completed=!!item.completed;item.writable=true;return item})
        compare(app.tasks().map(function(t){return t.title}).join("|"),
                "Soonest home task|Morning task|Afternoon task|Later work task|Undated task")
        app.taskList="work"
        compare(app.tasks().map(function(t){return t.title}).join("|"),"Morning task|Later work task")
        app.taskList=""
        app.showCompleted=true
        compare(app.tasks()[0].title,"Finished task")
    }
    function event(title,start,end,allDay) {
        return {kind:"VEVENT",id:title,title:title,start:start,end:end,all_day:!!allDay,
            calendar:"test",calendar_name:"Personal",writable:true,completed:false,pending:false,description:"",location:"",revision:"one"}
    }
    function test_timeline_duration_overlap_and_midnight() {
        var entries=Geometry.layout([
            event("First","2026-11-10T09:00","2026-11-10T10:00"),
            event("Overlap","2026-11-10T09:30","2026-11-10T11:00"),
            event("After","2026-11-10T11:00","2026-11-10T11:30"),
            event("Overnight","2026-11-09T23:00","2026-11-10T01:00"),
            event("Tomorrow","2026-11-10T23:00","2026-11-11T01:00"),
            event("All day","2026-11-10T00:00","2026-11-11T00:00",true)
        ],"2026-11-10")
        compare(entries.length,5)
        compare(entries[0].start,0);compare(entries[0].end,60)
        compare(entries[1].start,540);compare(entries[1].end,600)
        compare(entries[1].columns,2);compare(entries[2].columns,2)
        verify(entries[1].column!==entries[2].column)
        compare(entries[3].columns,1)
        compare(entries[4].end,1440)
        compare(Geometry.hit(entries,0.1,550).title,"First")
        compare(Geometry.hit(entries,0.8,580).title,"Overlap")
        compare(Geometry.hit(entries,0.8,700),null)
    }
    function test_swipe_calendar_views() {
        var timeline=findChild(app,"timelinePage")
        for (var i=0;i<2;++i) {
            app.changeView(i===0?"week":"day");wait(30)
            var gesture=findChild(timeline,"timelineGridGesture")
            var before=app.iso(app.selected)
            mouseDrag(gesture,700,timeline.scroll.contentY+200,-320,0,Qt.LeftButton,Qt.NoModifier,50)
            compare(app.iso(app.selected),app.iso(app.add(app.date(before),i===0?7:1)))
            wait(30)
            mouseDrag(gesture,400,timeline.scroll.contentY+200,320,0,Qt.LeftButton,Qt.NoModifier,50)
            compare(app.iso(app.selected),before)
        }
        app.changeView("month");wait(30)
        var month=findChild(app,"monthGesture")
        mouseDrag(month,700,200,-320,0,Qt.LeftButton,Qt.NoModifier,50)
        compare(app.selected.getMonth(),11)
        mouseDrag(month,400,200,320,0,Qt.LeftButton,Qt.NoModifier,50)
        compare(app.selected.getMonth(),10)
    }
    function test_vertical_scroll_does_not_navigate() {
        app.changeView("day");wait(30)
        var timeline=findChild(app,"timelinePage")
        var before=app.iso(app.selected), y=timeline.scroll.contentY
        mouseDrag(timeline.gridGesture,400,y+500,0,-240,Qt.LeftButton,Qt.NoModifier,50)
        compare(app.iso(app.selected),before)
        verify(timeline.scroll.contentY>y)
        compare(app.draft,null)
    }
    function test_tap_event_and_edge_back() {
        app.changeView("day");wait(30)
        var timeline=findChild(app,"timelinePage")
        mouseClick(timeline.gridGesture,300,(9.25-timeline.startHour)*timeline.hourHeight)
        verify(app.draft!==null)
        var before=app.iso(app.selected)
        var back=findChild(app,"eventBackGesture")
        mouseDrag(back,back.width-10,220,-300,0,Qt.LeftButton,Qt.NoModifier,50)
        compare(app.draft,null);compare(app.view,"day");compare(app.iso(app.selected),before)
        app.settingsOpen=true;wait(10)
        back=findChild(app,"settingsBackGesture")
        mouseDrag(back,back.width-10,220,-300,0,Qt.LeftButton,Qt.NoModifier,50)
        compare(app.settingsOpen,false)
    }
    function test_hold_empty_slot_creates_at_time() {
        app.changeView("day");wait(30)
        var timeline=findChild(app,"timelinePage")
        mousePress(timeline.gridGesture,300,(11.25-timeline.startHour)*timeline.hourHeight)
        wait(900)
        mouseRelease(timeline.gridGesture,300,(11.25-timeline.startHour)*timeline.hourHeight)
        verify(app.draft!==null)
        compare(app.draft.start,"2026-11-10T11:00")
        compare(app.draft.end,"2026-11-10T12:00")
        app.cancelEdit()
        compare(app.view,"day")
    }
    function test_header_and_month_taps_open_day() {
        var timeline=findChild(app,"timelinePage")
        var heading=findChild(timeline,"timelineHeadingsGesture")
        mouseClick(heading,4.5*timeline.dayWidth,30)
        compare(app.view,"day");compare(app.iso(app.selected),"2026-11-13")
        app.changeView("month");wait(30)
        var month=findChild(app,"monthGesture")
        mouseClick(month,2.5*month.width/7,2.5*month.height/6)
        compare(app.view,"day");compare(app.iso(app.selected),"2026-11-11")
    }
    function test_edge_back_ignores_short_or_vertical_drags_and_preserves_taps() {
        app.openItem(app.items[0]);wait(20)
        var back=findChild(app,"eventBackGesture")
        mouseDrag(back,40,300,-40,0,Qt.LeftButton,Qt.NoModifier,50)
        verify(app.draft!==null)
        mouseDrag(back,40,300,0,180,Qt.LeftButton,Qt.NoModifier,50)
        verify(app.draft!==null)
        // The edge overlaps part of Delete; ordinary taps must still reach that button.
        mouseClick(back,10,45)
        compare(app.confirmingDelete,true)
        app.cancelEdit()
    }
    function test_all_day_section_fits_each_view() {
        var timeline=findChild(app,"timelinePage"), events=[]
        for(var i=0;i<6;++i)events.push(event("Tuesday "+i,"2026-11-10T00:00","2026-11-11T00:00",true))
        for(var j=0;j<8;++j)events.push(event("Wednesday "+j,"2026-11-11T00:00","2026-11-12T00:00",true))
        app.items=events
        compare(timeline.allDayHeight,384)
        app.view="day";wait(30)
        compare(timeline.allDayHeight,288)
        var allDay=findChild(timeline,"allDayGesture")
        mouseClick(allDay,50,5*48+20)
        verify(app.draft!==null)
        compare(app.draft.title,"Tuesday 5")
        app.cancelEdit()
        app.selected=new Date(2026,10,12,12)
        compare(timeline.allDayHeight,42)
    }
    function test_hour_range_clips_events_without_changing_originals() {
        var early=event("Crosses 5am","2026-11-10T04:30","2026-11-10T06:00")
        var late=event("Crosses 10pm","2026-11-10T21:30","2026-11-10T23:00")
        var entries=Geometry.layout([
            event("Before","2026-11-10T03:00","2026-11-10T05:00"),early,late,
            event("After","2026-11-10T22:00","2026-11-10T23:00")
        ],"2026-11-10",300,1320)
        compare(entries.length,2)
        compare(entries[0].start,300);compare(entries[0].end,360)
        compare(entries[1].start,1290);compare(entries[1].end,1320)
        compare(entries[0].event.start,"2026-11-10T04:30")
        compare(entries[1].event.end,"2026-11-10T23:00")
        compare(Geometry.hit(entries,0.5,315).title,"Crosses 5am")
        compare(Geometry.hit(entries,0.5,1320),null)
    }
    function test_scheduled_and_priority_sort() {
        app.items=[
            {kind:"VTODO",title:"Later",scheduled:"2026-11-12",due:"2026-11-09",priority:1,calendar_name:"Work",writable:true,completed:false},
            {kind:"VTODO",title:"First",scheduled:"2026-11-10",due:"2026-11-20",priority:9,calendar_name:"Home",writable:true,completed:false},
            {kind:"VTODO",title:"Unscheduled",scheduled:"",due:"2026-11-01",priority:0,calendar_name:"Work",writable:true,completed:false}
        ]
        compare(app.taskSort,"scheduled")
        compare(app.tasks().map(function(t){return t.title}).join("|"),"First|Later|Unscheduled")
        app.savePreference("task_sort","priority")
        compare(app.tasks().map(function(t){return t.title}).join("|"),"Later|First|Unscheduled")
        app.savePreference("task_sort","due")
        compare(app.tasks().map(function(t){return t.title}).join("|"),"Unscheduled|Later|First")
    }
    function test_week_heading_and_time_formats() {
        compare(app.periodTitle(),"Nov 9–15, 2026")
        app.selected=new Date(2026,11,31,12)
        compare(app.periodTitle(),"Dec 28, 2026 – Jan 3, 2027")
        compare(Geometry.clock(0,"12h"),"12:00 AM")
        compare(Geometry.clock(720,"12h"),"12:00 PM")
        app.savePreference("time_format","12h")
        compare(app.formatTime("17:15"),"5:15 PM")
        app.savePreference("time_format","24h")
        compare(app.formatTime("17:15"),"17:15")
    }
    function test_time_dropdown_and_calendar_dropdown() {
        app.newItem();app.field=""
        var start=findChild(app,"startTimeDropdown")
        app.chooseTime(start,"start")
        var picker=findChild(app,"choicePicker")
        verify(picker.visible);compare(picker.options.length,96)
        compare(picker.options[37].value,"09:15")
        compare(app.field,"")
        picker.choose(37)
        compare(app.draft.start,"2026-11-10T09:15")
        var next=JSON.parse(JSON.stringify(app.status));next.collections.push({id:"second",name:"Work",events:true,tasks:true,writable:true,enabled:true});app.status=next
        app.chooseCalendar(findChild(app,"calendarDropdown"))
        compare(picker.options.length,2)
        picker.choose(1)
        compare(app.draft.calendar,"second");compare(app.draft.calendar_name,"Work")
    }
    function test_connection_password_masked_and_not_prefilled() {
        app.openSettings()
        compare(app.connectionDraft.app_password,"")
        verify(app.connectionDraft.has_password)
        app.field="connection.app_password"
        app.editText("test-secret")
        var keyboard=findChild(app,"touchKeyboard")
        verify(keyboard.secret)
        compare(findChild(keyboard,"keyboardInput").echoMode,TextInput.Password)
        compare(findChild(app,"passwordField").value,"••••••••")
        app.closeSettings()
        compare(app.connectionDraft.app_password,"")
    }
    function test_original_keyboard_letters_and_numbers() {
        app.newItem();wait(30)
        var keyboard=findChild(app,"touchKeyboard")
        compare(keyboard.height,663)
        compare(findChild(keyboard,"key_q").height,108)
        compare(findChild(keyboard,"key_q").width,findChild(keyboard,"key_a").width)
        compare(findChild(keyboard,"key_q").width,findChild(keyboard,"key_z").width)
        mouseClick(findChild(keyboard,"key_q"),20,20)
        compare(app.draft.title,"q")
        mouseClick(findChild(keyboard,"symbolModeButton"),20,20);wait(10)
        mouseClick(findChild(keyboard,"key_1"),20,20)
        compare(app.draft.title,"q1")
        mouseClick(findChild(keyboard,"key_/"),20,20)
        compare(app.draft.title,"q1/")
        verify(findChild(keyboard,"key_*")===null)
        keyboard.symbols=false
    }

    function test_partial_date_typing_never_inserts_time_separator() {
        app.newItem();app.field="date.start";wait(20)
        var keyboard=findChild(app,"touchKeyboard"),input=findChild(keyboard,"keyboardInput")
        input.selectAll();keyboard.type("2")
        compare(app.editValue(),"2");compare(input.text,"2")
        keyboard.type("026-12-03")
        compare(app.editValue(),"2026-12-03")
        compare(app.draft.start,"2026-12-03T09:00")
        compare(input.text.indexOf("T"),-1)
    }
    function test_hide_keyboard_preserves_partial_date_and_blocks_invalid_save() {
        app.newItem();app.field="date.start";app.editText("2026-0")
        var keyboard=findChild(app,"touchKeyboard")
        mouseClick(findChild(keyboard,"hideKeyboardButton"),30,15)
        compare(app.field,"");compare(app.dateValue("start"),"2026-0")
        app.save(false);compare(submitted,null);verify(app.message.indexOf("valid start")>=0)
        app.field="date.start";compare(app.editValue(),"2026-0")
        app.editText("2026-02-30");app.save(false);compare(submitted,null)
        app.cancelEdit()
    }
    function test_outside_tap_keeps_keyboard_open() {
        app.newItem();app.editText("Keep this title");wait(20)
        mouseClick(app,300,100)
        compare(app.field,"title")
        verify(findChild(app,"touchKeyboard").visible)
        compare(app.draft.title,"Keep this title")
        compare(submitted,null)
    }
}
