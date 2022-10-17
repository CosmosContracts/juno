#!/bin/bash

# create a function
function add_key {
    KEY_NAME=$1
    MNUMONIC=$2
    # echo "Adding key $KEY_NAME, $MNUMONIC"
    echo $MNUMONIC | junod keys add $KEY_NAME --recover --keyring-backend test 2> /dev/null
}

echo "Adding keys"
add_key val "blame tube add leopard fire next exercise evoke young team payment senior know estate mandate negative actual aware slab drive celery elevator burden utility"
add_key lo-test1 "notice oak worry limit wrap speak medal online prefer cluster roof addict wrist behave treat actual wasp year salad speed social layer crew genius"
add_key lo-test2 "quality vacuum heart guard buzz spike sight swarm shove special gym robust assume sudden deposit grid alcohol choice devote leader tilt noodle tide penalty"
add_key lo-test3 "symbol force gallery make bulk round subway violin worry mixture penalty kingdom boring survey tool fringe patrol sausage hard admit remember broken alien absorb"
add_key lo-test4 "bounce success option birth apple portion aunt rural episode solution hockey pencil lend session cause hedgehog slender journey system canvas decorate razor catch empty"
add_key lo-test5 "second render cat sing soup reward cluster island bench diet lumber grocery repeat balcony perfect diesel stumble piano distance caught occur example ozone loyal"
add_key lo-test6 "spatial forest elevator battle also spoon fun skirt flight initial nasty transfer glory palm drama gossip remove fan joke shove label dune debate quick"
add_key lo-test7 "noble width taxi input there patrol clown public spell aunt wish punch moment will misery eight excess arena pen turtle minimum grain vague inmate"
add_key lo-test8 "cream sport mango believe inhale text fish rely elegant below earth april wall rug ritual blossom cherry detail length blind digital proof identify ride"
add_key lo-test9 "index light average senior silent limit usual local involve delay update rack cause inmate wall render magnet common feature laundry exact casual resource hundred"
add_key lo-test10 "prefer forget visit mistake mixture feel eyebrow autumn shop pair address airport diesel street pass vague innocent poem method awful require hurry unhappy shoulder"
echo "Keys added..."