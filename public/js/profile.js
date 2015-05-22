/**
 * Created by gernest on 5/22/15.
 */

$(document).ready(function(){
    var tmpl=$('#profile-pic-upload');
    var dz=$('#my-pic').dropzone({
        url: "/uploads", // Set the url
        autoQueue: true,
        autoQueue: false,
        paramName: "profile",
        previewTemplate: tmpl.html(),
        clickable: "#profile-pic"
    });
    var dzGallery=$('#gallery-upload').dropzone({
        url: "/uploads", // Set the url
        autoQueue: true,
        autoQueue: false,
        paramName: "photos",
        previewTemplate: tmpl.html(),
        clickable: "#pandisha-kibao",
        previewsContainer: ".preview-container"

    });

});