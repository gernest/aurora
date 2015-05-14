/**
 * Created by gernest on 5/4/15.
 */

$(document).ready(function(){
    $('.button-collapse').sideNav();
    // profile
    var profileUpdate=function(){
        var upBtn=$('.up-btn');
        var upSection=$('.up-section');
        var upProgress=$('.up-progress');
        var upError=$('.up-error');
        var upForm=$('.up-form');
        var upMsg=$('.up-message');

        upSection.toggle();
        upProgress.toggle();
        upBtn.click(function(){
            upSection.toggle();
        });
        upForm.submit(function(e){
            upForm.toggle();
            upProgress.toggle();
            $.ajax({
                url:$(this).attr('action'),
                type:$(this).attr('method'),
                data:$(this).serialize(),
                success:function(data){
                    upProgress.toggle();
                    upMsg.html(data);
                },
                error:function(res,status,err){
                    upProgress.toggle();
                    upMsg.html(res.responseText);
                    upForm.toggle();
                }
            })
            e.preventDefault();
            e.unbind;
        });

    };
    profileUpdate();
});