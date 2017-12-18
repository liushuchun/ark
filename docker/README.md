##执行伪分布式
```
nohup python ../../tools/launch.py -n 1 -- --sync-dst-dir /workspace/tmp python train_imagenet.py  --data-train /workspace/train --data-val /workspace/val/ --gpus 0,1 --batch-size 128  --num-epochs 1  2>&1 &
```


##训练cifar
```
 python train_cifar10.py --network resnet --num-layers 110 --batch-size 128 --gpus 0 \
    --kv-store dist_device_sync
```


nohup python ../../tools/launch.py -n 4 --launcher ssh -H hosts  --sync-dst-dir /workspace/tmp python train_imagenet.py \
 --data-train /workspace/train --data-val /workspace/val/ --gpus 0,1 \
--batch-size 128 --num-epochs 1 2>&1 &