'use strict';
function PauzerCtrl($scope, $http){

    var form = $scope.form = {
        limit: 20,
        time: 1,
    };
    $scope.counter;
    var intervalVar;

    $http.get('state').success(function(data){
        if(data.limit != 0){
            $scope.form = {
                limit: data.limit,
                time: data.time,
            };
            startTimer(data.secondsLeft);
        }
    });


    $scope.setTime = function( ){
        $http.get('action/' + $scope.form.time + '/' + $scope.form.limit).success(function(data){ });
        startTimer( $scope.form.time * 60)
    };

    function startTimer(time){
        $scope.counter = time;
        clearInterval(intervalVar);
        intervalVar = setInterval(function(){
            $scope.$apply(function(){
                $scope.counter--;
                if( $scope.counter <= 0)
                    clearInterval(intervalVar);
            });
        }, 1000);
    }


}
PauzerCtrl.$inject = ['$scope', '$http'];
