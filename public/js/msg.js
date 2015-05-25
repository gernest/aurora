/**
 * Created by gernest on 5/14/15.
 */
$(document).ready(function(){
    var addr='ws://'+window.location.host+'/msg';
    var conn= new golem.Connection(addr,true);
    var fm=$('form#msg');
    var send=$('button.msg-send');
    var txt =$('textarea#msg-text');
    var pg=$('div.progress');
    var sendEvt='send';
    var alertSendSuccess = "sendSuccess";
    var alertSendFailed  = "sendFailled";
    var alertInbox       = "messageInbox";
    var msgAlert=$('#alert');

    send.click(function(e){
        e.preventDefault();
        pg.toggleClass('hide');
        var msg={
            'recepient_id':send.attr('msg-r'),
            'sender_id':send.attr('msg-s'),
            'text':txt.val()
        };
        conn.emit(sendEvt,msg);
        pg.toggleClass('hide');
        txt.val("");
    });
    conn.on(alertSendSuccess,function(data){
        Materialize.toast("ujumbe umefanikiwa kutumwa",900);
    });
    conn.on(alertSendFailed,function(data){
        console.log(data);
    });
    conn.on(alertInbox,function(data){
        msgAlert.text(Number(msgAlert.text()) +1);
        Materialize.toast("kuna ujumbe wako",999);
    });
});