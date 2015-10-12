$(function () {
    $('#startServer').on('click', function(){serverAction('start');});
    $('#stopServer').on('click', function(){serverAction('stop');});
    $('#autorestartServerCheckbox').on('change', autorestartServerCheckboxChange);
    serverStatusLongPoll(0);
});

function serverAction(action) {
    $.ajax({
        url: '/server',
        type: 'POST',
        dataType: 'json',
        data: {
            token: 'asdf',
            action: action
        },
        success: function (data, textStatus, jqXHR) {
            // success callback
            console.log("success");
        }
    }); 
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
