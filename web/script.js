$('#save-crm').on("submit", function(e) {
    e.preventDefault();
    send(
        $(this).attr('action'),
        formDataToObj($(this).serializeArray()),
        function () {
            return 0;
        }
    )
});

$("#save").on("submit", function(e) {
    e.preventDefault();
    send(
        $(this).attr('action'),
        formDataToObj($(this).serializeArray()),
        function () {
            return 0;
        }
    )
});

$("#add-bot").on("submit", function(e) {
    e.preventDefault();
    send(
        $(this).attr('action'),
        formDataToObj($(this).serializeArray()),
        function (data) {
            let bots = $("#bots");
            if (bots.hasClass("hide")) {
                bots.removeClass("hide")
            }
            $("#bots tbody").append(getBotTemplate(data));
        }
    )
});

$(document).on("click", ".activity-bot", function(e) {
    let but = $(this);
    send("/activity-bot/",
        {
            token: but.attr("data-token"),
            active: (but.attr("data-activity") === 'true'),
            clientId: $('input[name=clientId]').val(),
        },
        function () {
            if (but.attr("data-activity") === 'true') {
                but.find('i').replaceWith('<i class="material-icons">play_arrow</i>');
                but.attr("data-activity", "false")
            } else {
                but.find('i').replaceWith('<i class="material-icons">stop</i>');
                but.attr("data-activity", "true")
            }
        }
    )
});

function send(url, data, callback) {
    $('#msg').empty();
    $.ajax({
        url: url,
        data: JSON.stringify(data),
        type: "POST",
        success: callback,
        error: function (res){
            if (res.status < 400) {
                if (res.responseText) {
                    document.location.replace(
                        location.protocol.concat("//").concat(window.location.host) + res.responseText
                    );
                }
            } else {
                $('#msg').html(`<p class="err-msg truncate">${res.responseText}</p>`);
            }
        }
    });
}

function getBotTemplate(data) {
    let bot = JSON.parse(data);
    tmpl =
        `<tr>
            <td>${bot.name}</td>
            <td>${bot.token}</td>
            <td>
                <button class="activity-bot btn btn-meddium waves-effect waves-light red" type="submit" name="action"
                        data-activity="true" data-token="${bot.token}">
                    <i class="material-icons">stop</i>
                </button>
            </td>
        </tr>`;
    return tmpl;
}

function formDataToObj(formArray) {
    let obj = {};
    for (let i = 0; i < formArray.length; i++){
        obj[formArray[i]['name']] = formArray[i]['value'];
    }
    return obj;
}

$( document ).ready(function() {
    M.Tabs.init(document.getElementById("tab"));
    if ($("table tbody").children().length === 0) {
        $("#bots").addClass("hide");
    }
});
