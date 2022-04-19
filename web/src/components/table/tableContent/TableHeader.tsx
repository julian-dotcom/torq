type TableHeaderProps = {
  title: string;
};

export default function TableHeader({ title }: TableHeaderProps) {
  return <h1 className="text-2xl">{title}</h1>;
}
