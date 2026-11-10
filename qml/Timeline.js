// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
.pragma library

function minutes(value) {
    return Number(value.slice(11, 13)) * 60 + Number(value.slice(14, 16))
}

function clock(minute, format) {
    if (format === "12h") {
        var hour=Math.floor(minute/60)%24
        return String(hour%12||12)+":"+String(minute%60).padStart(2,"0")+(hour<12?" AM":" PM")
    }
    return String(Math.floor(minute / 60)).padStart(2, "0") + ":" + String(minute % 60).padStart(2, "0")
}

// Clip multi-day events to this day and allocate parallel columns for overlaps.
function layout(items, day, firstMinute, lastMinute) {
    if (firstMinute === undefined) firstMinute = 0
    if (lastMinute === undefined) lastMinute = 1440
    var entries = items.filter(function(e) {
        return e.kind === "VEVENT" && !e.all_day && e.start.slice(0, 10) <= day
            && (e.end > day + "T00:00" || e.start.slice(0, 10) === day)
    }).map(function(e) {
        var start = e.start.slice(0, 10) < day ? 0 : minutes(e.start)
        var end = e.end.slice(0, 10) > day ? 1440 : minutes(e.end)
        return {event:e, start:start, end:end, displayEnd:Math.min(1440, Math.max(end, start + 15)), column:0, columns:1}
    }).filter(function(e) {
        return e.start < lastMinute && (e.end > firstMinute || (e.start === e.end && e.start >= firstMinute))
    }).map(function(e) {
        e.start = Math.max(firstMinute, e.start)
        e.end = Math.min(lastMinute, e.end)
        e.displayEnd = Math.min(lastMinute, Math.max(e.end, e.start + 15))
        return e
    }).sort(function(a, b) {
        return a.start - b.start || b.end - a.end || a.event.title.localeCompare(b.event.title)
    })
    var group = [], occupied = [], groupEnd = -1, count = 0
    function finish() {
        group.forEach(function(e) { e.columns = count })
        group = []; occupied = []; count = 0
    }
    entries.forEach(function(e) {
        if (e.start >= groupEnd) finish()
        var column = 0
        while (column < occupied.length && occupied[column] > e.start) ++column
        occupied[column] = e.displayEnd
        e.column = column
        count = Math.max(count, column + 1)
        group.push(e)
        groupEnd = Math.max(groupEnd, e.displayEnd)
    })
    finish()
    return entries
}

function hit(entries, fraction, minute) {
    for (var i = entries.length - 1; i >= 0; --i) {
        var e = entries[i]
        if (minute >= e.start && minute < e.displayEnd
                && fraction >= e.column / e.columns && fraction < (e.column + 1) / e.columns)
            return e.event
    }
    return null
}
