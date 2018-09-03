$('#save-crm').on("submit", function(e) {
    e.preventDefault();
    send(
        $(this).attr('action'),
        formDataToObj($(this).serializeArray()),
        function (data) {
            sessionStorage.setItem("createdMsg", data.message);

            document.location.replace(
                location.protocol.concat("//").concat(window.location.host) + data.url
            );
        }
    )
});

$("#save").on("submit", function(e) {
    e.preventDefault();
    send(
        $(this).attr('action'),
        formDataToObj($(this).serializeArray()),
        function (data) {
            M.toast({html: data.message});
        }
    )
});

$("#add-bot").on("submit", function(e) {
    e.preventDefault();
    send(
        $(this).attr('action'),
        {
            connectionId: parseInt($(this).find('input[name=connectionId]').val()),
            token: $(this).find('input[name=token]').val(),
        },
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

$(document).on("click", ".delete-bot", function(e) {
    e.preventDefault();
    var but = $(this);
    var confirmText = JSON.parse(sessionStorage.getItem("confirmText"));

    $.confirm({
        title: false,
        content: confirmText["text"],
        useBootstrap: false,
        boxWidth: '30%',
        type: 'blue',
        backgroundDismiss: false,
        backgroundDismissAnimation: 'shake',
        buttons: {
            confirm: {
                text: confirmText["confirm"],
                action: function () {
                    send("/delete-bot/",
                        {
                            token: but.attr("data-token"),
                            connectionId: parseInt($('input[name=connectionId]').val()),
                        },
                        function () {
                            but.parents("tr").remove();
                            if ($("#bots tbody tr").length === 0) {
                                $("#bots").addClass("hide");
                            }
                        }
                    )
                },
            },
            cancel: {
                text: confirmText["cancel"],
            },
        }
    });
});

function send(url, data, callback) {
    $.ajax({
        url: url,
        data: JSON.stringify(data),
        type: "POST",
        success: callback,
        error: function (res){
            if (res.status >= 400) {
                M.toast({html: res.responseJSON.error})
            }
        }
    });
}

function getBotTemplate(data) {
    // let bot = JSON.parse(data);
    tmpl =
        `<tr>
            <td>${data.name}</td>
            <td>${data.token}</td>
            <td>
                <button class="delete-bot btn btn-small waves-effect waves-light light-blue darken-1" type="submit" name="action"
                        data-token="${data.token}">
                    <i class="material-icons">delete</i>
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

    if (!sessionStorage.getItem("confirmText")) {
        let confirmText = {};

        switch (navigator.language.split('-')[0]) {
            case "ru":
                confirmText["text"] = "Вы уверены, что хотите удалить?";
                confirmText["confirm"] = "да";
                confirmText["cancel"] = "нет";
                break;
            case "es":
                confirmText["text"] = "¿Estás seguro que quieres borrar?";
                confirmText["confirm"] = "sí";
                confirmText["cancel"] = "no";
                break;
            default:
                confirmText["text"] = "Are you sure you want to delete?";
                confirmText["confirm"] = "yes";
                confirmText["cancel"] = "no";
        }

        sessionStorage.setItem("confirmText", JSON.stringify(confirmText));
    }

    let createdMsg = sessionStorage.getItem("createdMsg");
    if (createdMsg) {
        setTimeout(function() {
            M.toast({html: createdMsg});
            sessionStorage.removeItem("createdMsg");
        }, 1000);
    }
});
