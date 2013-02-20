angular.module('pauzerFilters', []).filter('timeFormat', function() {
    return function( input ){
        if (!input)
            return
        function zeroPad(num, places) {
          var zero = places - num.toString().length + 1;
          return Array(+(zero > 0 && zero)).join("0") + num;
        }
        min = Math.floor(input/60);
        sec = zeroPad(input % 60, 2);
        return min + ':' + sec;
    };
});
