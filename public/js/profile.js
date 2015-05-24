/**
 * Created by gernest on 5/22/15.
 */

$(document).ready(function(){
    var tmpl=$('#profile-pic-upload');
    var dz=new Dropzone('#my-pic',{
        url: "/uploads", // Set the url
        autoQueue: true,
        paramName: "profile",
        previewTemplate: tmpl.html(),
        clickable: "#profile-pic",
        addRemoveLinks:true,
        maxFilesize:2,
        mazThumnailFileSize:2,
        thumbnailWidth:120,
        thumbnailHeigh:120,
        maxFiles:1,
        acceptedFiles:"image/jpg,image/png,image/jpeg",
        previewsContainer: "#pic-preview"
    });
    dz.on('complete',function(file){
        dz.removeFile(file);
    });
    dz.on('success',function(file,data){
        src='/imgs?'+data.query
        $('#profile-picture').attr('src',src);
        console.log(data);
    });
    var dzGallery=$('#gallery-upload').dropzone({
        url: "/uploads", // Set the url
        autoQueue: false,
        paramName: "photos",
        previewTemplate: tmpl.html(),
        clickable: "#pandisha-kibao",
        previewsContainer: ".preview-container"

    });
    $('#birth-date').pickadate({
        selectYears:true,
        selectMonths:true
    });
});