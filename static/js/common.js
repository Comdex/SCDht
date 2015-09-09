$(function() {
    document.onkeydown = function(e) {
        var ev = document.all ? window.event : e;
        if (ev.keyCode == 13) {
            $('#search').click()
        }
    }
    $('#search').click(function() {
        if ($('#key').val() == '') {
            $('.input-group').addClass('has-error');
        } else {
            window.location = '/search/' + encodeURIComponent($('#key').val())
        }
    });
    $('#gotop').hide();
    $('#gotop').click(function() {
        $('html,body').animate({
            scrollTop: '0px'
        }, 800)
    });
    $(window).bind('scroll', function() {
        var scrollTop = $(window).scrollTop();
        if (scrollTop > 100) $('#gotop').show();
        else $('#gotop').hide()
    })
});

function changeLanguage(lang) {
    var date = new Date();
    var expireDays = 9999999;
    date.setTime(date.getTime()+expireDays*24*3600*1000);
    document.cookie = "lang="+lang+";path=/; expires="+date.toGMTString();
    window.location.reload();
}
