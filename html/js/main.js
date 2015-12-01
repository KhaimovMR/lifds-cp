var baseBannedAccountFirstTr = '<tr class="danger"><td class="character"></td><td class="steam-id"></td></tr>';
var baseBannedAccountSecondTr = '<tr class="danger"><td class="character"></td></tr>';
var baseActiveAccountFirstTr = '<tr class="success"><td class="character"></td><td class="steam-id"></td></tr>';
var baseActiveAccountSecondTr = '<tr class="success"><td class="character"></td></tr>';
var baseOnlineCharacterTr = '<tr class="success"><td class="character"></td><td class="online-time"></td></tr>';

$(function () {
    $('#startServer').on('click', function(){serverAction('start');});
    $('#stopServer').on('click', function(){serverAction('stop');});
    $('#restartServer').on('click', function(){serverAction('restart');});
    $('#deleteTrees').on('click', function(){serverAction('delete-trees');});
    $('#deleteStubs').on('click', function(){serverAction('delete-stubs');});
    $('#autorestartServerCheckbox').on('change', autorestartServerCheckboxChange);
    $('#nav_tabs a[data-toggle="tab"]').on('show.bs.tab', function (e) {
        onlineCharactersTab(e.target);
        activeAccountsTab(e.target);
        bannedAccountsTab(e.target);
    });
    serverStatusLongPoll(0);
});

function onlineCharactersTab(linkElement) {
    if ($(linkElement).attr("href") === "#online_characters") {
        $(".online-characters-table").css("display: none;");
        serverAction("get-online-characters", {}, renderOnlineCharacters);
    }
}

function activeAccountsTab(linkElement) {
    if ($(linkElement).attr("href") === "#active_accounts") {
        $(".active-accounts-table").css("display: none;");
        serverAction("get-active-accounts", {}, renderAccounts);
    }
}

function bannedAccountsTab(linkElement) {
    if ($(linkElement).attr("href") === "#banned_accounts") {
        $(".banned-accounts-table").css("display: none;");
        serverAction("get-banned-accounts", {}, renderAccounts);
    }
}

function serverAction(action, params, successCallback) {
    if (typeof(params) === "undefined") {
        params = {};
    }

    if (typeof(successCallback) === "undefined") {
        successCallback = function() {};
    }

    $.ajax({
        url: '/server',
        type: 'POST',
        dataType: 'json',
        data: {
            token: 'asdf',
            action: action,
            params: params
        },
        success: function (data, textStatus, jqXHR) {
            successCallback(data, action);
            console.log("success");
        }
    }); 
}

function makeSteamUrl(steamId) {
    return '<a target="_blank" href="http://steamcommunity.com/profiles/' + steamId + '">' + steamId + '</a>';
}

function makeCharUrl(charId, charName) {
    return '<a rel="'+charId+'" class="character-link" href="#">' + charName + '</a>';
}

function renderOnlineCharacters(data, action) {
    var trHtml = baseOnlineCharacterTr;
    $(".online-characters-table tbody").html("");
    var tr;

    for (i in data) {
        tr = $(trHtml)

        tr.find(".character").html(makeCharUrl(data[i].ID,data[i].FullName));
        tr.find(".online-time").html(data[i].OnlineTime);
        $(".online-characters-table tbody").append(tr);
    }

    $('.character-link').on('click', onCharacterLinkClick);
    $(".online-characters-table").css("display: table;");
}

function renderAccounts(data, action) {
    var prefix = "";
    var firstTr = "";
    var secondTr = "";

    if (action === "get-banned-accounts") {
        prefix = "banned";
        firstTr = baseBannedAccountFirstTr;
        secondTr = baseBannedAccountSecondTr;
    } else if (action === "get-active-accounts") {
        prefix = "active";
        firstTr = baseActiveAccountFirstTr;
        secondTr = baseActiveAccountSecondTr;
    }

    $("." + prefix + "-accounts-table tbody").html("");

    for (i in data) {
        if (typeof(data[i].SteamID) === "undefined" || typeof(data[i].Characters) === "undefined") {
            break;
        }

        var charsCount = data[i].Characters.length
        var tr;

        for (j in data[i].Characters) {
            if (j === "0") {
                tr = $(firstTr);
                tr.find(".steam-id").attr('rowspan', charsCount)
                tr.find(".steam-id").html(makeSteamUrl(data[i].SteamID));
            } else {
                tr = $(secondTr);
            }

            tr.find(".character").html(makeCharUrl(data[i].Characters[j].ID,data[i].Characters[j].FullName));
            $("." + prefix + "-accounts-table tbody").append(tr);
        }
    }

    $('.character-link').on('click', onCharacterLinkClick);
    $("." + prefix + "-accounts-table").css("display: table;");
}

function serverStatusLongPoll(topic_version) {
    give_snapshot = typeof(give_snapshot) === 'undefined' ? false : true;

    $.ajax({
        url: '/server/status',
        type: 'POST',
        dataType: 'json',
        data: {
            token: 'asdf',
            topic_version: topic_version
        },
        timeout: 30000,
        success: function (data, textStatus, jqXHR) {
            $(".server-status.value").text(data.status);

            if (data.status.search('DOWN') != -1) {
                $(".server-status").removeClass('success');
                $(".server-status").removeClass('warning');
                $(".server-status").addClass('danger');
            } else if (data.status.search('UP') != -1) {
                $(".server-status").removeClass('danger');
                $(".server-status").removeClass('warning');
                $(".server-status").addClass('success');
            } else {
                $(".server-status").removeClass('danger');
                $(".server-status").removeClass('success');
                $(".server-status").addClass('warning');
            }

            if (data.online_statistics_enabled == true) {
                $(".online-characters-button").show();
            } else {
                $(".dashboard-button a").tab('show');
                $(".online-characters-button").hide();
            }

            setTimeout(function(){serverStatusLongPoll(data.topic_version);}, 0);
            return
        },
        error: function (jqXHR, textStatus, errorThrown) {
            milliseconds = errorThrown == "timeout" ? 0 : 3000;
            setTimeout(function(){serverStatusLongPoll(topic_version);}, milliseconds);
        }
    }); 
}

function autorestartServerCheckboxChange() {
    console.log("autocomplete");
}

function renderSkills(data, action) {
    for (var i in data) {
        $("#skill" + i).find(".value").html(data[i]);
    }
}

function renderDeathLog(data, action) {
    var html = "";

    for (var i in data) {
        html += data[i] + "<br/>";
    }

    $("#deathLogTab").html(html);
}

function renderOnlineHistory(data, action) {
    var html = "";
    html += "Total time (hours): " + data.TotalOnlineTime + "<br/><br/>";
    

    for (var i in data.History) {
        html += data.History[i] + "<br/>";
    }

    $("#onlineHistoryTab").html(html);
}

function onCharacterLinkClick(e) {
    var charId = $(this).attr("rel");
    var charName = $(this).text();
    $("#charSkillsModal").find(".value").html("0");
    serverAction("get-character-skills", {char_id: charId}, renderSkills);
    serverAction("get-character-death-log", {char_id: charId}, renderDeathLog);
    serverAction("get-character-online-history", {char_id: charId}, renderOnlineHistory);
    $("#charSkillsModal").find("h5.character-name").html(charName);
    $("#charSkillsModal").modal("show");
    $("#skillsTabLink").tab("show");

    return false;
}
