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
        function (data) {
            M.toast({html: data});
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
            $("#token").val("");
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
                but.find('i').replaceWith('<i class="material-icons">pause</i>');
                but.attr("data-activity", "true")
            }
        }
    )
});

function send(url, data, callback) {
    $.ajax({
        url: url,
        data: JSON.stringify(data),
        type: "POST",
        success: callback,
        error: function (res){
            if (res.status < 400) {
                if (res.responseText) {
                    let resObj = JSON.parse(res.responseText);
                    localStorage.setItem("createdMsg", resObj.Message);

                    document.location.replace(
                        location.protocol.concat("//").concat(window.location.host) + resObj.Url
                    );
                }
            } else {
                M.toast({html: res.responseText})
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
                <button class="activity-bot btn btn-small waves-effect waves-light light-blue darken-1" type="submit" name="action"
                        data-activity="true" data-token="${bot.token}">
                    <i class="material-icons">pause</i>
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

    let createdMsg = localStorage.getItem("createdMsg");
    if (createdMsg) {
        setTimeout(function() {
            M.toast({html: createdMsg});
            localStorage.removeItem("createdMsg");
        }, 1000);
    }
});
