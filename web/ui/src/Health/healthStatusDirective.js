
/* healthStatusDirective
 * directive for displaying health of a service/instance
 * using popover details
 */
(function() {
    'use strict';

    angular.module('healthStatus', [])
    .directive("healthStatus", [ "$translate",
    function($translate) {
        var linker = function($scope, element, attrs) {
            // cache some DOM elements
            var $el = $(element);
            var oldContent;

            // Called when the status of the host is updated.
            function update(){
                var service = $scope.service;
                var vhost = $scope.vhost;
                var statusObj = service.status;
                var popoverObj;
                var rollup = statusObj.statusRollup;
                var content = undefined;

                // If the service isn't running..
                if (!vhost.Enabled || rollup.bad > 0 || rollup.down > 0) {
                    content = $translate.instant("vhost_unavailable");
                } else if (service.desiredState === 0) {
                    content = $translate.instant("application") + " " + service.status.description;
                }

                // If there's no content to show in the popup or if the
                // status hasn't changed, exit.
                if ((!content) || oldContent === content){
                    oldContent = content;
                    return;
                }

                // Cache the popover content.
                oldContent = content;

                // NOTE: directly accessing the bootstrap popover data object here.
                popoverObj = $el.data("bs.popover");

                // if popover element already exists, update it
                if(popoverObj) {
                    // update the content
                    popoverObj.options.content = content;

                    // force popover to update using the new options
                    popoverObj.setContent();

                    // if the popover is currently visible, update
                    // it immediately, but turn off animation to
                    // prevent it fading in
                    if(popoverObj.$tip.is(":visible")){
                        popoverObj.options.animation = false;
                        popoverObj.show();
                        popoverObj.options.animation = true;
                    }
                // if popover element doesn't exist, create it
                } else {
                    // Set the popup with the content data.
                    $el.popover({
                        trigger: "hover",
                        placement: "top",
                        delay: 0,
                        content: content,
                        html: true,
                    });
                }
            }

            // if status object updates, update icon
            $scope.$watch("service.status", update);
        };

        return {
            restrict: "E",
            link: linker,
            scope: {
                // status object generated by serviceHealth
                vhost: "=",
                service: "="
            }
        };

    }]);
})();
