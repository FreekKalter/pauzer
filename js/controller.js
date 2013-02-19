'use strict';
function PauzerCtrl($scope, $http){
    var form = $scope.form = {
        limit: 100,
        time: 15,
    };

    $scope.counter;

    var intervalVar;

    $scope.setTime = function( ){
        $http.get('action/' + $scope.form.time + '/' + $scope.form.limit).success(function(data){ });
               
        $scope.counter = $scope.form.time * 60;
        clearInterval(intervalVar);
        intervalVar = setInterval(function(){
            $scope.$apply(function(){
                $scope.counter--;
            });
        }, 1000);
    };
}
PauzerCtrl.$inject = ['$scope', '$http'];
