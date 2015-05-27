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
    var chatBox=$('#i-chat');
    var idPrefix='aurora';
    var msgBox=''+
        '<li id="'+idPrefix+'{{=sender_id}}'+'">'+
          '<div class="collapsible-header"><i class="mdi-notification-sms blue-text notice">1</i>{{=sender_name}}</div>'+
          '<div class="collapsible-body">'+
            '<div class="row msgs">'+
              '<div class="msg-a">'+
                '<p>{{=text}}</p>'+
              '</div>'+
            '</div>'+
            '<div id="msg">'+
              '<div class="row">'+
                '<div class="progress hide">'+
                  '<div class="indeterminate"></div>'+
                '</div>'+
                '<form class="col s12" id="i-msg">'+
                  '<div class="row">'+
                    '<div class="input-field col s12">'+
                      '<textarea id="i-msg-text" class="materialize-textarea"></textarea>'+
                      '<label for="msg-text">Andika ujumbe wako hapa</label>'+
                    '</div>'+
                  '<div class="row">'+
                    '<button class="btn i-msg-send" msg-r="{{=recepient_id}}" msg-s="{{=sender_id}}">   jibu    </button>'+
                  '</div>'+
               '</form>'+
              '</div>'+
           '</div>'+
          '</div>'+
         '</li>';
    var singleMsg=''+
            '<div class="msg-a">'+
                '<p>{{=text}}</p>'+
            '</div>';

    var msgTmpl= new t(msgBox);
    var singleMsgTmpl=new t(singleMsg);
    var addInbox=function(obj){
        sid='#'+idPrefix+obj.sender_id;
        var base=chatBox.find(sid);
        if(base.length>0){
            base.find('.msgs').append(singleMsgTmpl.render(obj));
            n=base.find('i.notice');
            n.text(Number(n.text())+1);
        }else{
            chatBox.append(msgTmpl.render(obj))
                .one('click',function(e){
                    $(this).collapsible();
                    b =$(this);
                    $(this).find('i.notice').text('');
                    $(this).find('button.i-msg-send')
                        .on('click',function(e){
                            e.preventDefault();
                            var mfm=$(this).parents('form#i-msg');
                            var mtxt=mfm.find('textarea#i-msg-text');
                            var msg={
                                "recepient_id":$(this).attr('msg-s'),
                                "sender_id":$(this).attr('msg-r'),
                                "text":mtxt.val()
                            };
                            conn.emit(sendEvt,msg);
                            var msgB=''+
                                '<div class="msg-b">'+
                                    '<p>'+mtxt.val()+'</p>'+
                                '</div>';
                            b.find('.msgs').append(msgB);
                        });
                });
        }

    };

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
        addInbox(data);
    });
});