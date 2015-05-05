/**
 * Created by gernest on 5/4/15.
 */

$(document).ready(function(){
    $('.button-collapse').sideNav();

    // profile
    profileUpdateBtn=$('.update-btn');
    profileUpdateSect=$('.update-section');
    profileUpdateBtn.click(function(){
        profileUpdateSect.toggle();
    })



});