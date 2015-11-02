var baseBannedAccountFirstTr = '<tr class="danger"><td class="steam-id"></td><td class="character"></td></tr>';
var baseBannedAccountSecondTr = '<tr class="danger"><td class="character"></td></tr>';
var baseActiveAccountFirstTr = '<tr class="success"><td class="steam-id"></td><td class="character"></td></tr>';
var baseActiveAccountSecondTr = '<tr class="success"><td class="character"></td></tr>';

$(function () {
    $('#startServer').on('click', function(){serverAction('start');});
    $('#stopServer').on('click', function(){serverAction('stop');});
    $('#deleteTrees').on('click', function(){serverAction('delete-trees');});
    $('#deleteStubs').on('click', function(){serverAction('delete-stubs');});
    $('#autorestartServerCheckbox').on('change', autorestartServerCheckboxChange);
    $('#nav_tabs a[data-toggle="tab"]').on('show.bs.tab', function (e) {
        activeAccountsTab(e.target);
        bannedAccountsTab(e.target);
    });
    serverStatusLongPoll(0);
});

function activeAccountsTab(linkElement) {
    if ($(linkElement).attr("href") === "#active_accounts") {
        $(".active-accounts-table").css("display: none;");
        serverAction("get-active-accounts", renderAccounts);
    }
}

function bannedAccountsTab(linkElement) {
    if ($(linkElement).attr("href") === "#banned_accounts") {
        $(".banned-accounts-table").css("display: none;");
        serverAction("get-banned-accounts", renderAccounts);
    }
}

function serverAction(action, successCallback) {
    if (typeof(successCallback) === "undefined") {
        successCallback = function() {};
    }

    if (typeof(successCallbackParams) === "undefined") {
        successCallbackParams = {};
    }

    $.ajax({
        url: '/server',
        type: 'POST',
        dataType: 'json',
        data: {
            token: 'asdf',
            action: action
        },
        success: function (data, textStatus, jqXHR) {
            successCallback(data, action);
            console.log("success");
        }
    }); 
}

function makeSteamUrl(steamId) {
    return '<a href="http://steamcommunity.com/profiles/' + steamId + '">' + steamId + '</a>';
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

            tr.find(".character").html(data[i].Characters[j]);
            $("." + prefix + "-accounts-table tbody").append(tr);
        }
    }

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

            $(".server-version.value").text(data.current_version + " / " + data.available_version);
            
            if (data.current_version != data.available_version) {  
                $(".server-version").removeClass('success');
                $(".server-version").addClass('danger');
            } else {
                $(".server-version").removeClass('danger');
                $(".server-version").addClass('success');
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
